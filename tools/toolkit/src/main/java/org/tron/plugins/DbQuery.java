package org.tron.plugins;

import static org.tron.plugins.utils.Constant.ACCOUNT_STORE;
import static org.tron.plugins.utils.Constant.ALLOW_OLD_REWARD_OPT;
import static org.tron.plugins.utils.Constant.BLOCK_INDEX_STORE;
import static org.tron.plugins.utils.Constant.BLOCK_PRODUCED_INTERVAL;
import static org.tron.plugins.utils.Constant.BLOCK_STORE;
import static org.tron.plugins.utils.Constant.CHANGE_DELEGATION;
import static org.tron.plugins.utils.Constant.CURRENT_CYCLE_NUMBER;
import static org.tron.plugins.utils.Constant.DECIMAL_OF_VI_REWARD;
import static org.tron.plugins.utils.Constant.DELEGATION_STORE;
import static org.tron.plugins.utils.Constant.DYNAMIC_PROPERTY_STORE;
import static org.tron.plugins.utils.Constant.LATEST_BLOCK_HEADER_NUMBER;
import static org.tron.plugins.utils.Constant.MAINTENANCE_TIME_INTERVAL;
import static org.tron.plugins.utils.Constant.NEW_REWARD_ALGORITHM_EFFECTIVE_CYCLE;
import static org.tron.plugins.utils.Constant.REWARDS_KEY;
import static org.tron.plugins.utils.Constant.VOTES_ALL_WITNESSES;
import static org.tron.plugins.utils.Constant.VOTES_STORE;
import static org.tron.plugins.utils.Constant.VOTES_WITNESS_LIST;
import static org.tron.plugins.utils.Constant.WITNESS_STORE;

import com.google.common.collect.Maps;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.typesafe.config.Config;
import com.typesafe.config.ConfigFactory;
import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.concurrent.Callable;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.collections4.CollectionUtils;
import org.bouncycastle.util.encoders.Hex;
import org.rocksdb.RocksDBException;
import org.tron.common.utils.ByteArray;
import org.tron.common.utils.Commons;
import org.tron.common.utils.Pair;
import org.tron.common.utils.Sha256Hash;
import org.tron.common.utils.StringUtil;
import org.tron.core.capsule.AccountCapsule;
import org.tron.core.capsule.BlockCapsule;
import org.tron.core.capsule.VotesCapsule;
import org.tron.core.capsule.WitnessCapsule;
import org.tron.core.exception.BadItemException;
import org.tron.core.store.DelegationStore;
import org.tron.plugins.utils.FileUtils;
import org.tron.plugins.utils.JsonFormat;
import org.tron.plugins.utils.db.DBInterface;
import org.tron.plugins.utils.db.DBIterator;
import org.tron.plugins.utils.db.DbTool;
import org.tron.protos.Protocol.Transaction.Contract.ContractType;
import org.tron.protos.contract.WitnessContract.VoteWitnessContract;
import picocli.CommandLine;
import picocli.CommandLine.Command;

@Slf4j(topic = "query")
@Command(name = "query",
    description = "query the latest vote and reward information from the database.",
    exitCodeListHeading = "Exit Codes:%n",
    exitCodeList = {
        "0:Successful",
        "n:Internal error: exception occurred, please check logs/toolkit.log"})
public class DbQuery implements Callable<Integer> {

  @CommandLine.Spec
  CommandLine.Model.CommandSpec spec;

  @CommandLine.Option(names = {"-d", "--database-directory"},
      defaultValue = "output-directory",
      description = "java-tron database directory path. Default: ${DEFAULT-VALUE}")
  private String database;

  @CommandLine.Option(names = {"-c", "--config"},
      defaultValue = "query.conf",
      description = "config the votes and reward query options."
          + " Default: ${DEFAULT-VALUE}")
  private String config;

  @CommandLine.Option(names = {"-h", "--help"})
  private boolean help;

  private DBInterface witnessStore;
  private DBInterface votesStore;
  private DBInterface dynamicPropertiesStore;
  private DBInterface blockIndexStore;
  private DBInterface blockStore;
  private DBInterface accountStore;
  private DBInterface delegationStore;

  boolean allWitness = false;
  List<String> witnessList = new ArrayList<>();
  private Map<ByteString, VotesCapsule> voters = new HashMap<>();
  private Map<ByteString, VoteWitnessTx> votesTx = new HashMap<>();

  List<String> rewardAddressList = new ArrayList<>();
  Map<String, BigInteger> latestWitnessVi = new HashMap<>();

  private void initStore() throws IOException, RocksDBException {
    String srcDir = database + File.separator + "database";
    witnessStore = DbTool.getDB(srcDir, WITNESS_STORE);
    votesStore = DbTool.getDB(srcDir, VOTES_STORE);
    dynamicPropertiesStore = DbTool.getDB(srcDir, DYNAMIC_PROPERTY_STORE);
    blockIndexStore = DbTool.getDB(srcDir, BLOCK_INDEX_STORE);
    blockStore = DbTool.getDB(srcDir, BLOCK_STORE);
    accountStore = DbTool.getDB(srcDir, ACCOUNT_STORE);
    delegationStore = DbTool.getDB(srcDir, DELEGATION_STORE);
  }


  @Override
  public Integer call() throws Exception {
    if (help) {
      spec.commandLine().usage(System.out);
      return 0;
    }

    Config queryConfig;
    File file = Paths.get(config).toFile();
    if (file.exists() && file.isFile()) {
      queryConfig = ConfigFactory.parseFile(file);
    } else {
      logger.error("Query config file [" + config + "] not exists!");
      spec.commandLine().getErr().format("Fork config file: %s not exists!", config).println();
      return 1;
    }

    File dbFile = Paths.get(database).toFile();
    if (!dbFile.exists() || !dbFile.isDirectory()) {
      logger.error("Database [" + database + "] not exists!");
      spec.commandLine().getErr().format("Database %s not exists!", database).println();
      return 1;
    }
    File tmp = Paths.get(database, "database", "tmp").toFile();
    if (tmp.exists()) {
      FileUtils.deleteDir(tmp);
    }

    initStore();
    processVotes(queryConfig);
    processRewards(queryConfig);

    DbTool.close();
    return 0;
  }

  private void processVotes(Config queryConfig) throws BadItemException {
    if (queryConfig.hasPath(VOTES_ALL_WITNESSES)) {
      allWitness = queryConfig.getBoolean(VOTES_ALL_WITNESSES);
    }
    if (queryConfig.hasPath(VOTES_WITNESS_LIST)) {
      witnessList = queryConfig.getStringList(VOTES_WITNESS_LIST);
    }

    if (!allWitness && witnessList.isEmpty()) {
      spec.commandLine().getOut()
          .println("skip the vote query.");
      logger.info("skip the vote query.");
      return;
    }

    Map<ByteString, WitnessCapsule> witnesses = new HashMap<>();
    Map<ByteString, Long> oldWitnessCnt = new HashMap<>();
    DBIterator iterator = witnessStore.iterator();
    WitnessCapsule witnessCapsule;
    for (iterator.seekToFirst(); iterator.valid(); iterator.next()) {
      witnessCapsule = new WitnessCapsule(iterator.getValue());
      witnesses.put(ByteString.copyFrom(iterator.getKey()), witnessCapsule);
      oldWitnessCnt.put(witnessCapsule.getAddress(), witnessCapsule.getVoteCount());
    }

    Map<ByteString, Long> countWitness = countVote();
    loadVotesTx();

    AtomicInteger cnt = new AtomicInteger();
    votesTx.forEach((address, voteWitnessTx) -> {
      if (allWitness || existInWitnessList(voters.get(address), witnessList)) {
        cnt.getAndIncrement();
      }
    });
    spec.commandLine().getOut()
        .format("There are  %d related system-contract vote txs", cnt.get())
        .println();
    logger.info("There are {} related system-contract vote txs", cnt.get());

    cnt.set(-1);
    votesTx.forEach((address, voteWitnessTx) -> {
      VotesCapsule votesCapsule = voters.get(address);
      if (!allWitness && !existInWitnessList(votesCapsule, witnessList)) {
        return;
      }
      cnt.getAndIncrement();
      String txHashStr = voteWitnessTx.txHash.toString();
      spec.commandLine().getOut().format("tx %d: %s", cnt.get(), txHashStr).println();
      logger.info("tx {}: {}", cnt.get(), txHashStr);
//      String voteWitnessStr = JsonFormat.printToString(voteWitnessTx.voteWitnessContract, true);
//      spec.commandLine().getOut().println(voteWitnessStr);
//      logger.info(voteWitnessStr);
      String voteCapsuleStr = JsonFormat.printToString(votesCapsule.getInstance(), true);
      spec.commandLine().getOut().println(voteCapsuleStr);
      logger.info(voteCapsuleStr);
      voters.remove(address);
    });

    cnt.set(0);
    voters.forEach((address, votesCapsule) -> {
      if ((allWitness || existInWitnessList(voters.get(address), witnessList))
          && votesTx.get(address) == null) {
        cnt.getAndIncrement();
      }
    });

    spec.commandLine().getOut()
        .format("There are  %d related smart-contract votes txs", cnt.get())
        .println();
    logger.info("There are {} related smart-contract votes txs", cnt.get());
    cnt.set(-1);
    voters.forEach((address, voteCapsule) -> {
      if (!allWitness && !existInWitnessList(voteCapsule, witnessList)) {
        return;
      }
      cnt.getAndIncrement();
      spec.commandLine().getOut().format("tx %d", cnt.get()).println();
      logger.info("tx {}:", cnt.get());
      String voteCapsuleStr = JsonFormat.printToString(voteCapsule.getInstance(), true);
      spec.commandLine().getOut().println(voteCapsuleStr);
      logger.info(voteCapsuleStr);
    });

    countWitness.forEach((address, voteCount) -> {
      WitnessCapsule witness = witnesses.get(address);
      if (witness == null) {
        return;
      }
      witness.setVoteCount(witness.getVoteCount() + voteCount);
      witnesses.put(address, witness);
    });

    spec.commandLine().getOut().println("Display the witness list with latest vote count: ");
    logger.info("Display the witness list with latest vote count: ");
    List<WitnessCapsule> witnessCapsuleList;
    if (allWitness) {
      witnessCapsuleList = new ArrayList<>(witnesses.values());
      Collections.sort(witnessCapsuleList,
          Comparator.comparingLong(WitnessCapsule::getVoteCount).reversed());
    } else {
      List<WitnessCapsule> finalWitnessCapsuleList = new ArrayList<>();
      witnessList.forEach(
          witness -> {
            ByteString address = ByteString.copyFrom(Commons.decodeFromBase58Check(witness));
            WitnessCapsule witnessTmp = witnesses.get(address);
            if (witnessTmp == null) {
              spec.commandLine().getErr().format("address: %s is not witness", witness).println();
              logger.error("address: {} is not witness", witness);
              return;
            }
            finalWitnessCapsuleList.add(witnessTmp);
          }
      );
      witnessCapsuleList = finalWitnessCapsuleList;
    }

//      WitnessList.Builder builder = WitnessList.newBuilder();
//      witnessCapsuleList.forEach(witness -> builder.addWitnesses(witness.getInstance()));
//      String witnessesJson = JsonFormat.printToString(builder.build(), true);
//      spec.commandLine().getOut().println(witnessesJson);
//      logger.info(witnessesJson);
    cnt.set(-1);
    witnessCapsuleList.forEach(
        witness -> {
          cnt.getAndIncrement();
          String output = formatWitness(witness, oldWitnessCnt.get(witness.getAddress()));
          spec.commandLine().getOut().println(cnt.get() + " " + output);
          logger.info(cnt.get() + " " + output);
        });
  }

  private String formatWitness(WitnessCapsule witnessCapsule, long previousCnt) {
    return String.format("{\"address\": %s,"
            + "\"oldVoteCount\": %d,\"newVoteCount\": %d,"
            + "\"voteCountIncrement\": %d,\"url\": %s,"
            + "\"totalProduced\": %d,\"totalMissed\": %d,\"latestBlockNum\": %d,"
            + "\"latestSlotNum\": %d,\"isJobs\": %b}",
        StringUtil.encode58Check(witnessCapsule.getAddress().toByteArray()),
        previousCnt,
        witnessCapsule.getVoteCount(),
        witnessCapsule.getVoteCount() - previousCnt,
        witnessCapsule.getUrl(),
        witnessCapsule.getTotalProduced(),
        witnessCapsule.getTotalMissed(),
        witnessCapsule.getLatestBlockNum(),
        witnessCapsule.getLatestSlotNum(),
        witnessCapsule.getIsJobs());
  }


  private Map<ByteString, Long> countVote() {
    final Map<ByteString, Long> countWitness = Maps.newHashMap();
    DBIterator dbIterator = votesStore.iterator();
    long sizeCount = 0;
    dbIterator.seekToFirst();
    while (dbIterator.hasNext()) {
      Entry<byte[], byte[]> next = dbIterator.next();
      VotesCapsule votes = new VotesCapsule(next.getValue());
      voters.put(ByteString.copyFrom(next.getKey()), votes);

      votes.getOldVotes().forEach(vote -> {
        ByteString voteAddress = vote.getVoteAddress();
        long voteCount = vote.getVoteCount();
        if (countWitness.containsKey(voteAddress)) {
          countWitness.put(voteAddress, countWitness.get(voteAddress) - voteCount);
        } else {
          countWitness.put(voteAddress, -voteCount);
        }
      });
      votes.getNewVotes().forEach(vote -> {
        ByteString voteAddress = vote.getVoteAddress();
        long voteCount = vote.getVoteCount();
        if (countWitness.containsKey(voteAddress)) {
          countWitness.put(voteAddress, countWitness.get(voteAddress) + voteCount);
        } else {
          countWitness.put(voteAddress, voteCount);
        }
      });
      sizeCount++;
    }
    spec.commandLine().getOut().format("There are total %d new votes in this epoch", sizeCount)
        .println();
    logger.info("There are total {} new votes in this epoch", sizeCount);
    return countWitness;
  }

  private void loadVotesTx() throws BadItemException {
    long maintenanceTimeInterval = ByteArray
        .toLong(dynamicPropertiesStore.get(MAINTENANCE_TIME_INTERVAL));
    long maintenanceBlockCnt = maintenanceTimeInterval / BLOCK_PRODUCED_INTERVAL;
    long latestBlockNumber = ByteArray
        .toLong(dynamicPropertiesStore.get(LATEST_BLOCK_HEADER_NUMBER));
    long startBlock =
        latestBlockNumber > maintenanceBlockCnt ? latestBlockNumber - maintenanceBlockCnt : 0;
    long block;
    for (block = startBlock; block <= latestBlockNumber; block++) {
      byte[] blockHash = blockIndexStore.get(ByteArray.fromLong(block));
      BlockCapsule blockCapsule = new BlockCapsule(blockStore.get(blockHash));

      blockCapsule.getTransactions().forEach(txCapsule -> {
        ContractType txType = txCapsule.getInstance().getRawData().getContract(0).getType();
        if (!txType.equals(ContractType.VoteWitnessContract)) {
          return;
        }
        try {
          VoteWitnessContract voteWitnessContract = txCapsule.getInstance().getRawData()
              .getContract(0)
              .getParameter().unpack(VoteWitnessContract.class);
          ByteString ownerAddress = voteWitnessContract.getOwnerAddress();
          voteWitnessContract.getVotesList().forEach(vote -> {
            if (voters.keySet().contains(ownerAddress)) {
              votesTx.put(ownerAddress,
                  new VoteWitnessTx(txCapsule.getTransactionId(), voteWitnessContract));
            }
          });
        } catch (InvalidProtocolBufferException e) {
          e.printStackTrace();
          System.exit(-1);
        }
      });
    }
  }

  private class VoteWitnessTx {

    private Sha256Hash txHash;
    private VoteWitnessContract voteWitnessContract;

    VoteWitnessTx(Sha256Hash txHash, VoteWitnessContract voteWitnessContract) {
      this.txHash = txHash;
      this.voteWitnessContract = voteWitnessContract;
    }

    private Boolean existInWitnessList(List<String> witnessList) {
      AtomicBoolean exist = new AtomicBoolean(false);
      voteWitnessContract.getVotesList().forEach(vote -> {
        if (witnessList.contains(StringUtil.encode58Check(vote.getVoteAddress().toByteArray()))) {
          exist.set(true);
        }
      });
      return exist.get();
    }
  }

  private Boolean existInWitnessList(VotesCapsule votesCapsule, List<String> witnessList) {
    AtomicBoolean exist = new AtomicBoolean(false);
    votesCapsule.getOldVotes().forEach(vote -> {
      if (witnessList.contains(StringUtil.encode58Check(vote.getVoteAddress().toByteArray()))) {
        exist.set(true);
      }
    });
    votesCapsule.getNewVotes().forEach(vote -> {
      if (witnessList.contains(StringUtil.encode58Check(vote.getVoteAddress().toByteArray()))) {
        exist.set(true);
      }
    });
    return exist.get();
  }

  private void processRewards(Config config) {
    if (config.hasPath(REWARDS_KEY)) {
      rewardAddressList = config.getStringList(REWARDS_KEY);
    }
    if (rewardAddressList.isEmpty()) {
      spec.commandLine().getOut()
          .println("skip the reward query.");
      logger.info("skip the reward query.");
      return;
    }
    spec.commandLine().getOut().println("\nBegin to query the reward");

    loadAccumulateWitnessVi();
    rewardAddressList.forEach(address -> {
      long reward = queryReward(Commons.decodeFromBase58Check(address), false);
      long latestReward = queryReward(Commons.decodeFromBase58Check(address), true);
      spec.commandLine().getOut()
          .format("address: %s, cycle reward: %d, latest reward: %d, increment: %d", address,
              reward, latestReward, (latestReward - reward)).println();
      logger.info("address: {}, cycle reward: {}, latest reward: {}, increment: {}", address,
          reward, latestReward, (latestReward - reward));
    });

    spec.commandLine().getOut().println("Finish querying the reward");
  }

  private void loadAccumulateWitnessVi() {
    long cycle = getCurrentCycle();
    DBIterator iterator = witnessStore.iterator();
    WitnessCapsule witnessCapsule;
    for (iterator.seekToFirst(); iterator.valid(); iterator.next()) {
      witnessCapsule = new WitnessCapsule(iterator.getValue());
      accumulateWitnessVi(cycle, witnessCapsule.createDbKey(), witnessCapsule.getVoteCount());
    }
  }

  private void accumulateWitnessVi(long cycle, byte[] address, long voteCount) {
    BigInteger preVi = getWitnessVi(cycle - 1, address);
    long reward = getReward(cycle, address);
    if (reward == 0 || voteCount == 0) { // Just forward pre vi
      if (!BigInteger.ZERO.equals(preVi)) { // Zero vi will not be record
        setWitnessVi(cycle, address, preVi);
      }
    } else { // Accumulate delta vi
      BigInteger deltaVi = BigInteger.valueOf(reward)
          .multiply(DECIMAL_OF_VI_REWARD)
          .divide(BigInteger.valueOf(voteCount));
      setWitnessVi(cycle, address, preVi.add(deltaVi));
    }
  }

  private void setWitnessVi(long cycle, byte[] address, BigInteger value) {
    byte[] key = buildViKey(cycle, address);
    latestWitnessVi.put(ByteArray.toHexString(key), value);
  }

  private BigInteger getWitnessViFromMap(long cycle, byte[] address) {
    byte[] key = buildViKey(cycle, address);
    BigInteger value = latestWitnessVi.get(ByteArray.toHexString(key));
    if (value == null) {
      return BigInteger.ZERO;
    } else {
      return value;
    }
  }

  private long queryReward(byte[] address, boolean isLatest) {
    if (ByteArray.toLong(dynamicPropertiesStore.get(CHANGE_DELEGATION)) != 1) {
      return 0;
    }

    byte[] accountValue = accountStore.get(address);
    if (accountValue == null) {
      return 0;
    }
    AccountCapsule accountCapsule = new AccountCapsule(accountValue);
    long beginCycle = getBeginCycle(address);
    long endCycle = getEndCycle(address);
    long currentCycle = getCurrentCycle();

    long reward = 0;
    if (beginCycle > currentCycle) {
      return accountCapsule.getAllowance();
    }

    //withdraw the latest cycle reward
    if (beginCycle + 1 == endCycle && beginCycle < currentCycle) {
      AccountCapsule account = getAccountVote(beginCycle, address);
      if (account != null) {
        reward = computeReward(beginCycle, endCycle, account, false);
      }
      beginCycle += 1;
    }

    endCycle = currentCycle;
    if (CollectionUtils.isEmpty(accountCapsule.getVotesList())) {
      return reward + accountCapsule.getAllowance();
    }
    if ((!isLatest && beginCycle < endCycle) || (isLatest && beginCycle <= endCycle)) {
      reward += computeReward(beginCycle, endCycle, accountCapsule, isLatest);
    }
    return reward + accountCapsule.getAllowance();
  }

  private long getBeginCycle(byte[] address) {
    byte[] beginCycleValue = delegationStore.get(address);
    return beginCycleValue == null ? 0 : ByteArray.toLong(beginCycleValue);
  }

  private long getEndCycle(byte[] address) {
    byte[] endCycleValue = delegationStore.get(buildEndCycleKey(address));
    return endCycleValue == null ? -1L : ByteArray.toLong(endCycleValue);
  }

  private byte[] buildEndCycleKey(byte[] address) {
    return ("end-" + Hex.toHexString(address)).getBytes();
  }

  private long getCurrentCycle() {
    byte[] currentCycleValue = dynamicPropertiesStore.get(CURRENT_CYCLE_NUMBER);
    return currentCycleValue == null ? 0L : ByteArray.toLong(currentCycleValue);
  }

  private AccountCapsule getAccountVote(long cycle, byte[] address) {
    byte[] value = delegationStore.get(buildAccountVoteKey(cycle, address));
    if (value == null) {
      return null;
    } else {
      return new AccountCapsule(value);
    }
  }

  private byte[] buildAccountVoteKey(long cycle, byte[] address) {
    return (cycle + "-" + Hex.toHexString(address) + "-account-vote").getBytes();
  }

  private long getNewRewardAlgorithmEffectiveCycle() {
    byte[] value = dynamicPropertiesStore.get(NEW_REWARD_ALGORITHM_EFFECTIVE_CYCLE);
    return value == null ? 0L : ByteArray.toLong(value);
  }

  private Boolean allowOldRewardOpt() {
    byte[] value = dynamicPropertiesStore.get(ALLOW_OLD_REWARD_OPT);
    return value != null && ByteArray.toLong(value) == 1;
  }

  private BigInteger getWitnessVi(long cycle, byte[] address) {
    byte[] value = delegationStore.get(buildViKey(cycle, address));
    if (value == null) {
      return BigInteger.ZERO;
    } else {
      return new BigInteger(value);
    }
  }

  private byte[] buildViKey(long cycle, byte[] address) {
    return (cycle + "-" + Hex.toHexString(address) + "-vi").getBytes();
  }

  public long getReward(long cycle, byte[] address) {
    byte[] value = delegationStore.get(buildRewardKey(cycle, address));
    if (value == null) {
      return 0L;
    } else {
      return ByteArray.toLong(value);
    }
  }

  private byte[] buildRewardKey(long cycle, byte[] address) {
    return (cycle + "-" + Hex.toHexString(address) + "-reward").getBytes();
  }

  public long getWitnessVote(long cycle, byte[] address) {
    byte[] value = delegationStore.get(buildVoteKey(cycle, address));
    if (value == null) {
      return -1L;
    } else {
      return ByteArray.toLong(value);
    }
  }

  private byte[] buildVoteKey(long cycle, byte[] address) {
    return (cycle + "-" + Hex.toHexString(address) + "-vote").getBytes();
  }


  /**
   * Compute reward from begin cycle to end cycle, which endCycle must greater than beginCycle.
   * While computing reward after new reward algorithm taking effective cycle number, it will use
   * new algorithm instead of old way.
   *
   * @param beginCycle     begin cycle (include)
   * @param endCycle       end cycle (exclude)
   * @param accountCapsule account capsule
   * @param isLatest       whether to compute the latest reward
   * @return total reward
   */
  private long computeReward(long beginCycle, long endCycle, AccountCapsule accountCapsule,
      boolean isLatest) {
    if ((!isLatest && beginCycle >= endCycle) || (isLatest && beginCycle > endCycle)) {
      return 0;
    }

    long reward = 0;
    long newAlgorithmCycle = getNewRewardAlgorithmEffectiveCycle();
    List<Pair<byte[], Long>> srAddresses = accountCapsule.getVotesList().stream()
        .map(vote -> new Pair<>(vote.getVoteAddress().toByteArray(), vote.getVoteCount()))
        .collect(Collectors.toList());
    if (beginCycle < newAlgorithmCycle) {
      long oldEndCycle = Math.min(endCycle, newAlgorithmCycle);
      reward = getOldReward(beginCycle, oldEndCycle, srAddresses);
      beginCycle = oldEndCycle;
    }
    if ((!isLatest && beginCycle < endCycle) || (isLatest && beginCycle <= endCycle)) {
      for (Pair<byte[], Long> vote : srAddresses) {
        byte[] srAddress = vote.getKey();
        BigInteger beginVi = getWitnessVi(beginCycle - 1, srAddress);
        BigInteger endVi;
        if (!isLatest) {
          endVi = getWitnessVi(endCycle - 1, srAddress);
        } else {
          endCycle = getCurrentCycle();
          endVi = getWitnessViFromMap(endCycle, srAddress);
        }
        BigInteger deltaVi = endVi.subtract(beginVi);
        if (deltaVi.signum() <= 0) {
          continue;
        }
        long userVote = vote.getValue();
        reward += deltaVi.multiply(BigInteger.valueOf(userVote))
            .divide(DECIMAL_OF_VI_REWARD).longValue();
      }
    }
    return reward;
  }

  private long getNewRewardAlgorithmReward(long beginCycle, long endCycle,
      List<Pair<byte[], Long>> votes) {
    long reward = 0;
    if (beginCycle < endCycle) {
      for (Pair<byte[], Long> vote : votes) {
        byte[] srAddress = vote.getKey();
        BigInteger beginVi = getWitnessVi(beginCycle - 1, srAddress);
        BigInteger endVi = getWitnessVi(endCycle - 1, srAddress);
        BigInteger deltaVi = endVi.subtract(beginVi);
        if (deltaVi.signum() <= 0) {
          continue;
        }
        long userVote = vote.getValue();
        reward += deltaVi.multiply(BigInteger.valueOf(userVote))
            .divide(DECIMAL_OF_VI_REWARD).longValue();
      }
    }
    return reward;
  }

  private long computeReward(long cycle, List<Pair<byte[], Long>> votes) {
    long reward = 0;
    for (Pair<byte[], Long> vote : votes) {
      byte[] srAddress = vote.getKey();
      long totalReward = getReward(cycle, srAddress);
      if (totalReward <= 0) {
        continue;
      }
      long totalVote = getWitnessVote(cycle, srAddress);
      if (totalVote == DelegationStore.REMARK || totalVote == 0) {
        continue;
      }
      long userVote = vote.getValue();
      // Replace floating-point division with integer-based calculation
      reward += (userVote * totalReward) / totalVote;
    }
    return reward;
  }

  private long getOldReward(long begin, long end, List<Pair<byte[], Long>> votes) {
    if (allowOldRewardOpt()) {
      return getNewRewardAlgorithmReward(begin, end, votes);
    }
    long reward = 0;
    for (long cycle = begin; cycle < end; cycle++) {
      reward += computeReward(cycle, votes);
    }
    return reward;
  }

}
