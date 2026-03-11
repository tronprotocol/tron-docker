package org.tron.plugins;

import static java.lang.System.arraycopy;
import static org.tron.core.Constant.CREATE_ACCOUNT_TRANSACTION_MAX_BYTE_SIZE;
import static org.tron.core.Constant.CREATE_ACCOUNT_TRANSACTION_MIN_BYTE_SIZE;
import static org.tron.core.Constant.DYNAMIC_ENERGY_INCREASE_FACTOR_RANGE;
import static org.tron.core.config.Parameter.ChainConstant.ONE_YEAR_BLOCK_NUMBERS;
import static org.tron.plugins.utils.ChainParameters.CHAIN_PARAMETERS;
import static org.tron.plugins.utils.Constant.ACCOUNTS_KEY;
import static org.tron.plugins.utils.Constant.ACCOUNT_ADDRESS;
import static org.tron.plugins.utils.Constant.ACCOUNT_ASSET;
import static org.tron.plugins.utils.Constant.ACCOUNT_BALANCE;
import static org.tron.plugins.utils.Constant.ACCOUNT_NAME;
import static org.tron.plugins.utils.Constant.ACCOUNT_OWNER;
import static org.tron.plugins.utils.Constant.ACCOUNT_STORE;
import static org.tron.plugins.utils.Constant.ACCOUNT_TRC10_BALANCE;
import static org.tron.plugins.utils.Constant.ACCOUNT_TRC10_ID;
import static org.tron.plugins.utils.Constant.ACCOUNT_TYPE;
import static org.tron.plugins.utils.Constant.ACTIVE_WITNESSES;
import static org.tron.plugins.utils.Constant.ASSET_ISSUE_V2;
import static org.tron.plugins.utils.Constant.CONTRACT_STORE;
import static org.tron.plugins.utils.Constant.DYNAMIC_PROPERTY_STORE;
import static org.tron.plugins.utils.Constant.LATEST_BLOCK_HEADER_TIMESTAMP;
import static org.tron.plugins.utils.Constant.MAINTENANCE_TIME;
import static org.tron.plugins.utils.Constant.MAX_ACTIVE_WITNESS_NUM;
import static org.tron.plugins.utils.Constant.STORAGE_ROW_STORE;
import static org.tron.plugins.utils.Constant.TRC20_ACCOUNT;
import static org.tron.plugins.utils.Constant.TRC20_BALANCE;
import static org.tron.plugins.utils.Constant.TRC20_BALANCES_POSITION;
import static org.tron.plugins.utils.Constant.TRC20_CONTRACTS_KEY;
import static org.tron.plugins.utils.Constant.TRC20_CONTRACT_ADDRESS;
import static org.tron.plugins.utils.Constant.WITNESS_ADDRESS;
import static org.tron.plugins.utils.Constant.WITNESS_KEY;
import static org.tron.plugins.utils.Constant.WITNESS_SCHEDULE_STORE;
import static org.tron.plugins.utils.Constant.WITNESS_STORE;
import static org.tron.plugins.utils.Constant.WITNESS_URL;
import static org.tron.plugins.utils.Constant.WITNESS_VOTE;

import com.google.common.primitives.Bytes;
import com.google.common.primitives.Longs;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.typesafe.config.Config;
import com.typesafe.config.ConfigFactory;
import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.Callable;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.ArrayUtils;
import org.rocksdb.RocksDBException;
import org.tron.common.crypto.Hash;
import org.tron.common.utils.ByteArray;
import org.tron.common.utils.ByteUtil;
import org.tron.common.utils.Commons;
import org.tron.core.capsule.AccountCapsule;
import org.tron.core.capsule.StorageRowCapsule;
import org.tron.core.capsule.WitnessCapsule;
import org.tron.core.exception.ContractValidateException;
import org.tron.plugins.utils.ChainParameters;
import org.tron.plugins.utils.Constant;
import org.tron.plugins.utils.FileUtils;
import org.tron.plugins.utils.db.DBInterface;
import org.tron.plugins.utils.db.DBIterator;
import org.tron.plugins.utils.db.DbTool;
import org.tron.protos.Protocol.Account;
import org.tron.protos.Protocol.AccountType;
import org.tron.protos.Protocol.Permission;
import org.tron.protos.contract.SmartContractOuterClass.SmartContract;
import picocli.CommandLine;
import picocli.CommandLine.Command;

@Slf4j(topic = "fork")
@Command(name = "fork",
    description = "Modify the database of java-tron for shadow fork testing.",
    exitCodeListHeading = "Exit Codes:%n",
    exitCodeList = {
        "0:Successful",
        "n:Internal error: exception occurred, please check logs/toolkit.log"})
public class DbFork implements Callable<Integer> {

  @CommandLine.Spec
  CommandLine.Model.CommandSpec spec;

  @CommandLine.Option(names = {"-d", "--database-directory"},
      defaultValue = "output-directory",
      description = "java-tron database directory path. Default: ${DEFAULT-VALUE}")
  private String database;

  @CommandLine.Option(names = {"-c", "--config"},
      defaultValue = "fork.conf",
      description = "config the new witnesses, balances, etc for shadow fork."
          + " Default: ${DEFAULT-VALUE}")
  private String config;

  @CommandLine.Option(names = {"-r", "--retain-witnesses"},
      description = "retain the previous witnesses and active witnesses.")
  private boolean retain;

  @CommandLine.Option(names = {"-h", "--help"})
  private boolean help;

  private String srcDir;

  @Override
  public Integer call() throws IOException, RocksDBException, ContractValidateException {
    if (help) {
      spec.commandLine().usage(System.out);
      return 0;
    }

    File dbFile = Paths.get(database).toFile();
    if (!dbFile.exists() || !dbFile.isDirectory()) {
      logger.error("Database [" + database + "] not exists!");
      spec.commandLine().getErr().format("Database %s not exists!", database).println();
      System.exit(-1);
    }
    File tmp = Paths.get(database, "database", "tmp").toFile();
    if (tmp.exists()) {
      FileUtils.deleteDir(tmp);
    }

    Config forkConfig = ConfigFactory.load();
    File file = Paths.get(config).toFile();
    if (file.exists() && file.isFile()) {
      forkConfig = ConfigFactory.parseFile(file);
    } else {
      logger.error("Fork config file [" + config + "] not exists!");
      spec.commandLine().getErr().format("Fork config file: %s not exists!", config).println();
      System.exit(-1);
    }

    srcDir = database + File.separator + "database";
    processWitnesses(forkConfig);
    processAccounts(forkConfig);
    processTrc20Contracts(forkConfig);
    processDynamicProperties(forkConfig);

    DbTool.close();
    return 0;
  }

  private void processWitnesses(Config forkConfig) throws RocksDBException, IOException {
    DBInterface witnessStore = DbTool.getDB(srcDir, WITNESS_STORE);
    DBInterface witnessScheduleStore = DbTool.getDB(srcDir, WITNESS_SCHEDULE_STORE);

    if (!retain) {
      logger.info("Erase the previous witnesses and active witnesses.");
      spec.commandLine().getOut().println("Erase the previous witnesses and active witnesses.");
      witnessScheduleStore.delete(ACTIVE_WITNESSES);
      DBIterator iterator = witnessStore.iterator();
      for (iterator.seekToFirst(); iterator.valid(); iterator.next()) {
        witnessStore.delete(iterator.getKey());
      }
    } else {
      logger.warn("Keep the previous witnesses and active witnesses.");
      spec.commandLine().getOut().println("Keep the previous witnesses and active witnesses.");
    }

    if (!forkConfig.hasPath(WITNESS_KEY)) {
      return;
    }

    List<? extends Config> witnesses = forkConfig.getConfigList(WITNESS_KEY);
    if (witnesses.isEmpty()) {
      spec.commandLine().getOut().println("no witness listed in the config.");
    }
    witnesses = witnesses.stream()
        .filter(c -> c.hasPath(WITNESS_ADDRESS))
        .collect(Collectors.toList());

    if (witnesses.isEmpty()) {
      spec.commandLine().getOut().println("no witness listed in the config.");
    }

    List<ByteString> witnessList = new ArrayList<>();
    witnesses.stream().forEach(
        w -> {
          ByteString address = ByteString.copyFrom(
              Commons.decodeFromBase58Check(w.getString(WITNESS_ADDRESS)));
          WitnessCapsule witness = new WitnessCapsule(address);
          witness.setIsJobs(true);
          if (w.hasPath(WITNESS_VOTE) && w.getLong(WITNESS_VOTE) > 0) {
            witness.setVoteCount(w.getLong(WITNESS_VOTE));
          }
          if (w.hasPath(WITNESS_URL)) {
            witness.setUrl(w.getString(WITNESS_URL));
          }
          witnessStore.put(address.toByteArray(), witness.getData());
          witnessList.add(witness.getAddress());
        });

    witnessList.sort(Comparator.comparingLong((ByteString b) ->
            new WitnessCapsule(witnessStore.get(b.toByteArray())).getVoteCount())
        .reversed()
        .thenComparing(Comparator.comparingInt(ByteString::hashCode).reversed()));
    List<ByteString> activeWitnesses = witnessList.subList(0,
        witnesses.size() >= MAX_ACTIVE_WITNESS_NUM ? MAX_ACTIVE_WITNESS_NUM : witnessList.size());
    witnessScheduleStore.put(ACTIVE_WITNESSES, getActiveWitness(activeWitnesses));
    logger.info("{} witnesses and {} active witnesses have been modified.",
        witnesses.size(), activeWitnesses.size());
    spec.commandLine().getOut().format("%d witnesses and %d active witnesses have been modified.",
        witnesses.size(), activeWitnesses.size()).println();
  }

  private void processAccounts(Config forkConfig) throws RocksDBException, IOException {
    if (!forkConfig.hasPath(ACCOUNTS_KEY)) {
      return;
    }

    List<? extends Config> accounts = forkConfig.getConfigList(ACCOUNTS_KEY);
    if (accounts.isEmpty()) {
      spec.commandLine().getOut().println("no account listed in the config.");
    }

    DBInterface accountAssetStore = DbTool.getDB(srcDir, ACCOUNT_ASSET);
    accounts = accounts.stream()
        .filter(c -> c.hasPath(ACCOUNT_ADDRESS))
        .collect(Collectors.toList());

    if (accounts.isEmpty()) {
      spec.commandLine().getOut().println("no account listed in the config.");
    }

    DBInterface accountStore = DbTool.getDB(srcDir, ACCOUNT_STORE);
    DBInterface assetIssueV2Store = DbTool.getDB(srcDir, ASSET_ISSUE_V2);
    accounts.stream().forEach(
        a -> {
          byte[] address = Commons.decodeFromBase58Check(a.getString(ACCOUNT_ADDRESS));
          byte[] value = accountStore.get(address);
          Account account = null;
          try {
            account = ArrayUtils.isEmpty(value) ? null : Account.parseFrom(value);
          } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
            System.exit(-1);
          }

          if (Objects.isNull(account)) {
            ByteString byteAddress = ByteString.copyFrom(
                Commons.decodeFromBase58Check(a.getString(ACCOUNT_ADDRESS)));
            account = Account.newBuilder().setAddress(byteAddress).build();
          }
          AccountCapsule accountCapsule = new AccountCapsule(account);

          if (a.hasPath(ACCOUNT_BALANCE) && a.getLong(ACCOUNT_BALANCE) > 0) {
            accountCapsule.setBalance(a.getLong(ACCOUNT_BALANCE));
          }
          if (a.hasPath(ACCOUNT_NAME)) {
            accountCapsule.setAccountName(
                ByteArray.fromString(a.getString(ACCOUNT_NAME)));
          }
          if (a.hasPath(ACCOUNT_TYPE)) {
            accountCapsule.updateAccountType(
                AccountType.valueOf(a.getString(ACCOUNT_TYPE)));
          }

          if (a.hasPath(ACCOUNT_OWNER)) {
            byte[] owner = Commons.decodeFromBase58Check(a.getString(ACCOUNT_OWNER));
            Permission ownerPermission = AccountCapsule
                .createDefaultOwnerPermission(ByteString.copyFrom(owner));
            accountCapsule.updatePermissions(ownerPermission, null, null);
          }

          if (a.hasPath(ACCOUNT_TRC10_ID) && a.hasPath(ACCOUNT_TRC10_BALANCE)
              && a.getLong(ACCOUNT_TRC10_BALANCE) > 0) {
            String trc10Id = a.getString(ACCOUNT_TRC10_ID);
            if (assetIssueV2Store.get(ByteArray.fromString(trc10Id)) != null) {
              if (accountCapsule.getAssetOptimized()) {
                byte[] k = Bytes.concat(address, ByteArray.fromString(trc10Id));
                accountAssetStore.put(k, Longs.toByteArray(a.getLong(ACCOUNT_TRC10_BALANCE)));
              } else {
                Map<String, Long> assetMapV2 = new HashMap<>(account.getAssetV2Map());
                assetMapV2.put(trc10Id, a.getLong(ACCOUNT_TRC10_BALANCE));
                accountCapsule.clearAssetV2();
                accountCapsule.addAssetMapV2(assetMapV2);
              }
            } else {
              logger.info("TRC10: {} not exists in the database.", trc10Id);
              spec.commandLine().getOut().format("TRC10: %s not exists in the database.", trc10Id)
                  .println();
            }
          }

          accountStore.put(address, accountCapsule.getData());
        });
    logger.info("{} accounts have been modified.", accounts.size());
    spec.commandLine().getOut().format("%d accounts have been modified.", accounts.size())
        .println();
  }

  private void processTrc20Contracts(Config forkConfig) throws RocksDBException, IOException {
    if (!forkConfig.hasPath(TRC20_CONTRACTS_KEY)) {
      return;
    }
    List<? extends Config> trc20Contracts = forkConfig.getConfigList(TRC20_CONTRACTS_KEY);
    if (trc20Contracts.isEmpty()) {
      spec.commandLine().getOut().println("no TRC20 contract listed in the config.");
    }

    DBInterface storageRowStore = DbTool.getDB(srcDir, STORAGE_ROW_STORE);
    DBInterface contractStore = DbTool.getDB(srcDir, CONTRACT_STORE);
    trc20Contracts = trc20Contracts.stream()
        .filter(c -> c.hasPath(TRC20_CONTRACT_ADDRESS) && c.hasPath(TRC20_BALANCES_POSITION) &&
            c.hasPath(TRC20_ACCOUNT) && c.hasPath(TRC20_BALANCE))
        .collect(Collectors.toList());

    if (trc20Contracts.isEmpty()) {
      spec.commandLine().getOut().println("no TRC20 contract listed in the config.");
    }

    AtomicInteger cnt = new AtomicInteger();

    trc20Contracts.stream().forEach(
        contract -> {
          byte[] contractAddress = Commons
              .decodeFromBase58Check(contract.getString(TRC20_CONTRACT_ADDRESS));
          if (contractStore.get(contractAddress) == null) {
            spec.commandLine().getErr().format("TRC20 contract: %s not exists in the database.",
                contract.getString(TRC20_CONTRACT_ADDRESS)).println();
            return;
          }

          SmartContract smartContract;
          try {
            smartContract = SmartContract.parseFrom(contractStore.get(contractAddress));
          } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
            return;
          }

          int balancesSlotPosition = 0;
          if (contract.getInt(TRC20_BALANCES_POSITION) > 0) {
            balancesSlotPosition = contract.getInt(TRC20_BALANCES_POSITION);
          }
          byte[] addressWithPrefix = Commons
              .decodeFromBase58Check(contract.getString(TRC20_ACCOUNT));
          byte[] address = ByteArray.subArray(addressWithPrefix, 1, 21);
          String paddedAddress = String
              .format("%064x", new BigInteger(ByteArray.toHexString(address), 16));
          String paddedSlot = String.format("%064x", balancesSlotPosition);
          byte[] contractKey = Hash.sha3(ByteArray.fromHexString(
              paddedAddress + paddedSlot));

          byte[] addressHash;
          byte[] trxHash = smartContract.getTrxHash().toByteArray();
          if (ByteUtil.isNullOrZeroArray(trxHash)) {
            addressHash = Hash.sha3(contractAddress);
          } else {
            addressHash = Hash.sha3(ByteUtil.merge(contractAddress, trxHash));
          }

          int contractVersion = smartContract.getVersion();
          if (contractVersion == 1) {
            contractKey = Hash.sha3(contractKey);
          }
          byte[] rowKey = new byte[contractKey.length];
          arraycopy(addressHash, 0, rowKey, 0, 16);
          arraycopy(contractKey, 16, rowKey, 16, 16);

          String paddedBalance = String
              .format("%064x", new BigInteger(contract.getString(TRC20_BALANCE), 10));
          byte[] rowValue = ByteArray.fromHexString(paddedBalance);
          StorageRowCapsule storageRowCapsule = new StorageRowCapsule(rowKey, rowValue);

          storageRowStore.put(rowKey, storageRowCapsule.getData());
          cnt.getAndIncrement();
        });
    logger.info("{} TRC20 contracts have been modified.", cnt.get());
    spec.commandLine().getOut()
        .format("%d TRC20 contracts have been modified.", cnt.get())
        .println();
  }

  private void processDynamicProperties(Config forkConfig)
      throws RocksDBException, IOException, ContractValidateException {
    DBInterface dynamicPropertiesStore = DbTool.getDB(srcDir, DYNAMIC_PROPERTY_STORE);

    if (!forkConfig.hasPath(CHAIN_PARAMETERS) || forkConfig.getConfig(CHAIN_PARAMETERS).isEmpty()) {
      return;
    }

    Config chainParamsConfig = forkConfig.getConfig(CHAIN_PARAMETERS);

    if (chainParamsConfig.hasPath(ChainParameters.LATEST_BLOCK_TIMESTAMP)
        && chainParamsConfig.getLong(ChainParameters.LATEST_BLOCK_TIMESTAMP) > 0) {
      long latestBlockHeaderTimestamp = chainParamsConfig.getLong(
          ChainParameters.LATEST_BLOCK_TIMESTAMP);
      dynamicPropertiesStore
          .put(LATEST_BLOCK_HEADER_TIMESTAMP, ByteArray.fromLong(latestBlockHeaderTimestamp));
      logger.info("The latest block header timestamp has been modified as {}.",
          latestBlockHeaderTimestamp);
      spec.commandLine().getOut().format("The LATEST_BLOCK_HEADER_TIMESTAMP has been modified "
          + "as %d.", latestBlockHeaderTimestamp).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MAINTENANCE_INTERVAL)
        && chainParamsConfig.getLong(ChainParameters.MAINTENANCE_INTERVAL) > 0) {
      long maintenanceTimeInterval = chainParamsConfig.getLong(
          ChainParameters.MAINTENANCE_INTERVAL);
      if (maintenanceTimeInterval < 3 * 27 * 1000 || maintenanceTimeInterval > 24 * 3600 * 1000) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [3 * 27 * 1000,24 * 3600 * 1000]");
      }
      dynamicPropertiesStore
          .put("MAINTENANCE_TIME_INTERVAL".getBytes(), ByteArray.fromLong(maintenanceTimeInterval));
      logger.info("The maintenance time interval has been modified as {}.",
          maintenanceTimeInterval);
      spec.commandLine().getOut().format("The MAINTENANCE_TIME_INTERVAL has been modified as %d.",
          maintenanceTimeInterval).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.NEXT_MAINTENANCE_TIME)
        && chainParamsConfig.getLong(ChainParameters.NEXT_MAINTENANCE_TIME) > 0) {
      long nextMaintenanceTime = chainParamsConfig.getLong(ChainParameters.NEXT_MAINTENANCE_TIME);
      dynamicPropertiesStore.put(MAINTENANCE_TIME, ByteArray.fromLong(nextMaintenanceTime));
      logger.info("The NEXT_MAINTENANCE_TIME has been modified as {}.",
          nextMaintenanceTime);
      spec.commandLine().getOut().format("The NEXT_MAINTENANCE_TIME has been modified as %d.",
          nextMaintenanceTime).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ACCOUNT_UPGRADE_COST)) {
      long accountUpgradeCost = chainParamsConfig.getLong(ChainParameters.ACCOUNT_UPGRADE_COST);
      checkLongRange(accountUpgradeCost);
      dynamicPropertiesStore.put("ACCOUNT_UPGRADE_COST".getBytes(),
          ByteArray.fromLong(accountUpgradeCost));

      spec.commandLine().getOut().format("The ACCOUNT_UPGRADE_COST has been modified as %d.",
          accountUpgradeCost).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.CREATE_ACCOUNT_FEE)) {
      long createAccountFee = chainParamsConfig.getLong(ChainParameters.CREATE_ACCOUNT_FEE);
      checkLongRange(createAccountFee);
      dynamicPropertiesStore.put("CREATE_ACCOUNT_FEE".getBytes(),
          ByteArray.fromLong(createAccountFee));

      spec.commandLine().getOut().format("The CREATE_ACCOUNT_FEE has been modified as %d.",
          createAccountFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.TRANSACTION_FEE)) {
      long transactionFee = chainParamsConfig.getLong(ChainParameters.TRANSACTION_FEE);
      checkLongRange(transactionFee);
      dynamicPropertiesStore.put("TRANSACTION_FEE".getBytes(),
          ByteArray.fromLong(transactionFee));
      spec.commandLine().getOut().format("The TRANSACTION_FEE has been modified as %d.",
          transactionFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ASSET_ISSUE_FEE)) {
      long assetIssueFee = chainParamsConfig.getLong(ChainParameters.ASSET_ISSUE_FEE);
      checkLongRange(assetIssueFee);
      dynamicPropertiesStore.put("ASSET_ISSUE_FEE".getBytes(),
          ByteArray.fromLong(assetIssueFee));
      spec.commandLine().getOut().format("The ASSET_ISSUE_FEE has been modified as %d.",
          assetIssueFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.WITNESS_PAY_PER_BLOCK)) {
      long witnessPayPerBlock = chainParamsConfig.getLong(ChainParameters.WITNESS_PAY_PER_BLOCK);
      checkLongRange(witnessPayPerBlock);
      dynamicPropertiesStore.put("WITNESS_PAY_PER_BLOCK".getBytes(),
          ByteArray.fromLong(witnessPayPerBlock));
      spec.commandLine().getOut().format("The WITNESS_PAY_PER_BLOCK has been modified as %d.",
          witnessPayPerBlock).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.WITNESS_STAND_BY_ALLOWANCE)) {
      long witnessStandByAllowance = chainParamsConfig.getLong(
          ChainParameters.WITNESS_STAND_BY_ALLOWANCE);
      checkLongRange(witnessStandByAllowance);
      dynamicPropertiesStore.put("WITNESS_STANDBY_ALLOWANCE".getBytes(),
          ByteArray.fromLong(witnessStandByAllowance));
      spec.commandLine().getOut().format("The WITNESS_STANDBY_ALLOWANCE has been modified as %d.",
          witnessStandByAllowance).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.CREATE_NEW_ACCOUNT_FEE_IN_SYSTEM_CONTRACT)) {
      long createNewAccountFeeInSystemContract = chainParamsConfig.getLong(
          ChainParameters.CREATE_NEW_ACCOUNT_FEE_IN_SYSTEM_CONTRACT);
      checkLongRange(createNewAccountFeeInSystemContract);
      dynamicPropertiesStore.put("CREATE_NEW_ACCOUNT_FEE_IN_SYSTEM_CONTRACT".getBytes(),
          ByteArray.fromLong(createNewAccountFeeInSystemContract));
      spec.commandLine().getOut()
          .format("The CREATE_NEW_ACCOUNT_FEE_IN_SYSTEM_CONTRACT has been modified as %d.",
              createNewAccountFeeInSystemContract).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.CREATE_NEW_ACCOUNT_BANDWIDTH_RATE)) {
      long createNewAccountBandwidthRate = chainParamsConfig.getLong(
          ChainParameters.CREATE_NEW_ACCOUNT_BANDWIDTH_RATE);
      checkLongRange(createNewAccountBandwidthRate);
      dynamicPropertiesStore.put("CREATE_NEW_ACCOUNT_BANDWIDTH_RATE".getBytes(),
          ByteArray.fromLong(createNewAccountBandwidthRate));
      spec.commandLine().getOut()
          .format("The CREATE_NEW_ACCOUNT_BANDWIDTH_RATE has been modified as %d.",
              createNewAccountBandwidthRate).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_CREATION_OF_CONTRACTS)) {
      long allowCreationOfContracts = chainParamsConfig.getLong(
          ChainParameters.ALLOW_CREATION_OF_CONTRACTS);
      checkLongBoolean(allowCreationOfContracts);
      dynamicPropertiesStore.put("ALLOW_CREATION_OF_CONTRACTS".getBytes(),
          ByteArray.fromLong(allowCreationOfContracts));
      spec.commandLine().getOut()
          .format("The ALLOW_CREATION_OF_CONTRACTS has been modified as %d.",
              allowCreationOfContracts).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.REMOVE_THE_POWER_OF_THE_GR)) {
      long removeThePowerOfTheGr = chainParamsConfig.getLong(
          ChainParameters.REMOVE_THE_POWER_OF_THE_GR);
      checkLongNegative(removeThePowerOfTheGr);
      dynamicPropertiesStore.put("REMOVE_THE_POWER_OF_THE_GR".getBytes(),
          ByteArray.fromLong(removeThePowerOfTheGr));
      spec.commandLine().getOut()
          .format("The REMOVE_THE_POWER_OF_THE_GR has been modified as %d.",
              removeThePowerOfTheGr).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ENERGY_FEE)) {
      long energyFee = chainParamsConfig.getLong(
          ChainParameters.ENERGY_FEE);
      dynamicPropertiesStore.put("ENERGY_FEE".getBytes(),
          ByteArray.fromLong(energyFee));
      spec.commandLine().getOut().format("The ENERGY_FEE has been modified as %d.",
          energyFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.EXCHANGE_CREATE_FEE)) {
      long exchangeCreateFee = chainParamsConfig.getLong(
          ChainParameters.EXCHANGE_CREATE_FEE);
      dynamicPropertiesStore.put("EXCHANGE_CREATE_FEE".getBytes(),
          ByteArray.fromLong(exchangeCreateFee));
      spec.commandLine().getOut().format("The EXCHANGE_CREATE_FEE has been modified as %d.",
          exchangeCreateFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MAX_CPU_TIME_OF_ONE_TX)) {
      long maxCpuTimeOfOneTx = chainParamsConfig.getLong(
          ChainParameters.MAX_CPU_TIME_OF_ONE_TX);
      if (maxCpuTimeOfOneTx < 10 || maxCpuTimeOfOneTx > 400) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [10,400]");
      }
      dynamicPropertiesStore.put("MAX_CPU_TIME_OF_ONE_TX".getBytes(),
          ByteArray.fromLong(maxCpuTimeOfOneTx));
      spec.commandLine().getOut().format("The MAX_CPU_TIME_OF_ONE_TX has been modified as %d.",
          maxCpuTimeOfOneTx).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_UPDATE_ACCOUNT_NAME)) {
      long allowUpdateAccountName = chainParamsConfig.getLong(
          ChainParameters.ALLOW_UPDATE_ACCOUNT_NAME);
      checkLongBoolean(allowUpdateAccountName);
      dynamicPropertiesStore.put("ALLOW_UPDATE_ACCOUNT_NAME".getBytes(),
          ByteArray.fromLong(allowUpdateAccountName));
      spec.commandLine().getOut().format("The ALLOW_UPDATE_ACCOUNT_NAME has been modified as %d.",
          allowUpdateAccountName).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_SAME_TOKEN_NAME)) {
      long allowSameTokenName = chainParamsConfig.getLong(
          ChainParameters.ALLOW_SAME_TOKEN_NAME);
      checkLongBoolean(allowSameTokenName);
      dynamicPropertiesStore.put("ALLOW_SAME_TOKEN_NAME".getBytes(),
          ByteArray.fromLong(allowSameTokenName));
      spec.commandLine().getOut().format("The ALLOW_SAME_TOKEN_NAME has been modified as %d.",
          allowSameTokenName).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_DELEGATE_RESOURCE)) {
      long allowDelegateResource = chainParamsConfig.getLong(
          ChainParameters.ALLOW_DELEGATE_RESOURCE);
      checkLongBoolean(allowDelegateResource);
      dynamicPropertiesStore.put("ALLOW_DELEGATE_RESOURCE".getBytes(),
          ByteArray.fromLong(allowDelegateResource));
      spec.commandLine().getOut().format("The ALLOW_DELEGATE_RESOURCE has been modified as %d.",
          allowDelegateResource).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.TOTAL_ENERGY_LIMIT)) {
      long totalEnergyLimit = chainParamsConfig.getLong(
          ChainParameters.TOTAL_ENERGY_LIMIT);
      checkLongRange(totalEnergyLimit);
      dynamicPropertiesStore.put("TOTAL_ENERGY_LIMIT".getBytes(),
          ByteArray.fromLong(totalEnergyLimit));
      spec.commandLine().getOut().format("The TOTAL_ENERGY_LIMIT has been modified as %d.",
          totalEnergyLimit).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_TRANSFER_TRC10)) {
      long allowTvmTransferTrc10 = chainParamsConfig.getLong(
          ChainParameters.ALLOW_TVM_TRANSFER_TRC10);
      checkLongBoolean(allowTvmTransferTrc10);
      dynamicPropertiesStore.put("ALLOW_TVM_TRANSFER_TRC10".getBytes(),
          ByteArray.fromLong(allowTvmTransferTrc10));
      spec.commandLine().getOut().format("The ALLOW_TVM_TRANSFER_TRC10 has been modified as %d.",
          allowTvmTransferTrc10).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.TOTAL_ENERGY_CURRENT_LIMIT)) {
      long totalEnergyCurrentLimit = chainParamsConfig.getLong(
          ChainParameters.TOTAL_ENERGY_CURRENT_LIMIT);
      checkLongRange(totalEnergyCurrentLimit);
      dynamicPropertiesStore.put("TOTAL_ENERGY_LIMIT".getBytes(),
          ByteArray.fromLong(totalEnergyCurrentLimit));
      spec.commandLine().getOut().format("The TOTAL_ENERGY__CURRENT_LIMIT has been modified as %d.",
          totalEnergyCurrentLimit).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_MULTI_SIGN)) {
      long allowMultiSign = chainParamsConfig.getLong(
          ChainParameters.ALLOW_MULTI_SIGN);
      checkLongBoolean(allowMultiSign);
      dynamicPropertiesStore.put("ALLOW_MULTI_SIGN".getBytes(),
          ByteArray.fromLong(allowMultiSign));
      spec.commandLine().getOut().format("The ALLOW_MULTI_SIGN has been modified as %d.",
          allowMultiSign).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_ADAPTIVE_ENERGY)) {
      long allowAdaptiveEnergy = chainParamsConfig.getLong(
          ChainParameters.ALLOW_ADAPTIVE_ENERGY);
      checkLongBoolean(allowAdaptiveEnergy);
      dynamicPropertiesStore.put("ALLOW_ADAPTIVE_ENERGY".getBytes(),
          ByteArray.fromLong(allowAdaptiveEnergy));
      spec.commandLine().getOut().format("The ALLOW_ADAPTIVE_ENERGY has been modified as %d.",
          allowAdaptiveEnergy).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.TOTAL_ENERGY_TARGET_LIMIT)) {
      long totalEnergyTargetLimit = chainParamsConfig.getLong(
          ChainParameters.TOTAL_ENERGY_TARGET_LIMIT);
      dynamicPropertiesStore.put("TOTAL_ENERGY_TARGET_LIMIT".getBytes(),
          ByteArray.fromLong(totalEnergyTargetLimit));
      spec.commandLine().getOut().format("The TOTAL_ENERGY_TARGET_LIMIT has been modified as %d.",
          totalEnergyTargetLimit).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.TOTAL_ENERGY_AVERAGE_USAGE)) {
      long totalEnergyAverageUsage = chainParamsConfig.getLong(
          ChainParameters.TOTAL_ENERGY_AVERAGE_USAGE);
      dynamicPropertiesStore.put("TOTAL_ENERGY_AVERAGE_USAGE".getBytes(),
          ByteArray.fromLong(totalEnergyAverageUsage));
      spec.commandLine().getOut().format("The TOTAL_ENERGY_AVERAGE_USAGE has been modified as %d.",
          totalEnergyAverageUsage).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.UPDATE_ACCOUNT_PERMISSION_FEE)) {
      long allowUpdateEnergyLimit = chainParamsConfig.getLong(
          ChainParameters.UPDATE_ACCOUNT_PERMISSION_FEE);
      if (allowUpdateEnergyLimit < 0 || allowUpdateEnergyLimit > ChainParameters.MAX_SUPPLY) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [0," + ChainParameters.MAX_SUPPLY + "]");
      }
      dynamicPropertiesStore.put("UPDATE_ACCOUNT_PERMISSION_FEE".getBytes(),
          ByteArray.fromLong(allowUpdateEnergyLimit));
      spec.commandLine().getOut().format("The UPDATE_ACCOUNT_PERMISSION_FEE has been modified as %d.",
          allowUpdateEnergyLimit).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MULTI_SIGN_FEE)) {
      long multiSignFee = chainParamsConfig.getLong(
          ChainParameters.MULTI_SIGN_FEE);
      if (multiSignFee < 0 || multiSignFee > ChainParameters.MAX_SUPPLY) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [0," + ChainParameters.MAX_SUPPLY + "]");
      }
      dynamicPropertiesStore.put("MULTI_SIGN_FEE".getBytes(), ByteArray.fromLong(multiSignFee));
      spec.commandLine().getOut().format("The MULTI_SIGN_FEE has been modified as %d.",
          multiSignFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_ACCOUNT_STATE_ROOT)) {
      long allowAccountStateRoot = chainParamsConfig.getLong(
          ChainParameters.ALLOW_ACCOUNT_STATE_ROOT);
      checkLongBoolean(allowAccountStateRoot);
      dynamicPropertiesStore.put("ALLOW_ACCOUNT_STATE_ROOT".getBytes(),
          ByteArray.fromLong(allowAccountStateRoot));
      spec.commandLine().getOut().format("The ALLOW_ACCOUNT_STATE_ROOT has been modified as %d.",
          allowAccountStateRoot).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_PROTO_FILTER_NUM)) {
      long allowProtoFilterNum = chainParamsConfig.getLong(
          ChainParameters.ALLOW_PROTO_FILTER_NUM);
      checkLongBoolean(allowProtoFilterNum);
      dynamicPropertiesStore.put("ALLOW_PROTO_FILTER_NUM".getBytes(),
          ByteArray.fromLong(allowProtoFilterNum));
      spec.commandLine().getOut().format("The ALLOW_PROTO_FILTER_NUM has been modified as %d.",
          allowProtoFilterNum).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_CONSTANTINOPLE)) {
      long allowTVMConstantinople = chainParamsConfig.getLong(
          ChainParameters.ALLOW_TVM_CONSTANTINOPLE);
      checkLongBoolean(allowTVMConstantinople);
      dynamicPropertiesStore.put("ALLOW_TVM_CONSTANTINOPLE".getBytes(),
          ByteArray.fromLong(allowTVMConstantinople));
      spec.commandLine().getOut().format("The ALLOW_TVM_CONSTANTINOPLE has been modified as %d.",
          allowTVMConstantinople).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_SOLIDITY059)) {
      long allowTVMSolidity059 = chainParamsConfig.getLong(
          ChainParameters.ALLOW_TVM_SOLIDITY059);
      checkLongBoolean(allowTVMSolidity059);
      dynamicPropertiesStore.put("ALLOW_TVM_SOLIDITY_059".getBytes(),
          ByteArray.fromLong(allowTVMSolidity059));
      spec.commandLine().getOut().format("The ALLOW_TVM_SOLIDITY_059 has been modified as %d.",
          allowTVMSolidity059).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_ISTANBUL)) {
      long allowTVMIStanbul = chainParamsConfig.getLong(
          ChainParameters.ALLOW_TVM_ISTANBUL);
      checkLongBoolean(allowTVMIStanbul);
      dynamicPropertiesStore.put("ALLOW_TVM_ISTANBUL".getBytes(),
          ByteArray.fromLong(allowTVMIStanbul));
      spec.commandLine().getOut().format("The ALLOW_TVM_ISTANBUL has been modified as %d.",
          allowTVMIStanbul).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_SHIELDED_TRC20_TRANSACTION)) {
      long allowShieldedTRC20Transaction = chainParamsConfig.getLong(
          ChainParameters.ALLOW_SHIELDED_TRC20_TRANSACTION);
      checkLongBoolean(allowShieldedTRC20Transaction);
      dynamicPropertiesStore.put("ALLOW_SHIELDED_TRC20_TRANSACTION".getBytes(),
          ByteArray.fromLong(allowShieldedTRC20Transaction));
      spec.commandLine().getOut().format("The ALLOW_SHIELDED_TRC20_TRANSACTION has been modified as %d.",
          allowShieldedTRC20Transaction).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.FORBID_TRANSFER_TO_CONTRACT)) {
      long forbidTransferToContract = chainParamsConfig.getLong(
          ChainParameters.FORBID_TRANSFER_TO_CONTRACT);
      checkLongBoolean(forbidTransferToContract);
      dynamicPropertiesStore.put("FORBID_TRANSFER_TO_CONTRACT".getBytes(),
          ByteArray.fromLong(forbidTransferToContract));
      spec.commandLine().getOut().format("The FORBID_TRANSFER_TO_CONTRACT has been modified as %d.",
          forbidTransferToContract).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ADAPTIVE_RESOURCE_LIMIT_TARGET_RATIO)) {
      long adaptiveResourceLimitTargetRatio = chainParamsConfig.getLong(
          ChainParameters.ADAPTIVE_RESOURCE_LIMIT_TARGET_RATIO);
      if (adaptiveResourceLimitTargetRatio < 1 || adaptiveResourceLimitTargetRatio > 1_000) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [1,1_000]");
      }
      dynamicPropertiesStore.put("ADAPTIVE_RESOURCE_LIMIT_TARGET_RATIO".getBytes(),
          ByteArray.fromLong(adaptiveResourceLimitTargetRatio));
      spec.commandLine().getOut()
          .format("The ADAPTIVE_RESOURCE_LIMIT_TARGET_RATIO has been modified as %d.",
              adaptiveResourceLimitTargetRatio).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ADAPTIVE_RESOURCE_LIMIT_MULTIPLIER)) {
      long adaptiveResourceLimitMultiplier = chainParamsConfig.getLong(
          ChainParameters.ADAPTIVE_RESOURCE_LIMIT_MULTIPLIER);
      if (adaptiveResourceLimitMultiplier < 1 || adaptiveResourceLimitMultiplier > 10_000L) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [1,10_000]");
      }
      dynamicPropertiesStore.put("ADAPTIVE_RESOURCE_LIMIT_MULTIPLIER".getBytes(),
          ByteArray.fromLong(adaptiveResourceLimitMultiplier));
      spec.commandLine().getOut()
          .format("The ADAPTIVE_RESOURCE_LIMIT_MULTIPLIER has been modified as %d.",
              adaptiveResourceLimitMultiplier).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.CHANGE_DELEGATION)) {
      long changeDelegation = chainParamsConfig.getLong(
          ChainParameters.CHANGE_DELEGATION);
      checkLongBoolean(changeDelegation);
      dynamicPropertiesStore.put("CHANGE_DELEGATION".getBytes(),
          ByteArray.fromLong(changeDelegation));
      spec.commandLine().getOut().format("The CHANGE_DELEGATION has been modified as %d.",
          changeDelegation).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.WITNESS127_PAY_PER_BLOCK)) {
      long witness127PayPerBlock = chainParamsConfig.getLong(
          ChainParameters.WITNESS127_PAY_PER_BLOCK);
      checkLongRange(witness127PayPerBlock);
      dynamicPropertiesStore.put("WITNESS_127_PAY_PER_BLOCK".getBytes(),
          ByteArray.fromLong(witness127PayPerBlock));
      spec.commandLine().getOut().format("The WITNESS_127_PAY_PER_BLOCK has been modified as %d.",
          witness127PayPerBlock).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_MARKET_TRANSACTION)) {
      long allowMarketTransaction = chainParamsConfig.getLong(
          ChainParameters.ALLOW_MARKET_TRANSACTION);
      checkLongBoolean(allowMarketTransaction);
      dynamicPropertiesStore.put("ALLOW_MARKET_TRANSACTION".getBytes(),
          ByteArray.fromLong(allowMarketTransaction));
      spec.commandLine().getOut().format("The ALLOW_MARKET_TRANSACTION has been modified as %d.",
          allowMarketTransaction).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MARKET_SELL_FEE)) {
      long marketSellFee = chainParamsConfig.getLong(
          ChainParameters.MARKET_SELL_FEE);
      if (marketSellFee < 0 || marketSellFee > 10_000_000_000L) {
        throw new ContractValidateException(
            "Bad MARKET_SELL_FEE parameter value, valid range is [0,10_000_000_000L]");
      }
      dynamicPropertiesStore.put("MARKET_SELL_FEE".getBytes(),
          ByteArray.fromLong(marketSellFee));
      spec.commandLine().getOut().format("The MARKET_SELL_FEE has been modified as %d.",
          marketSellFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MARKET_CANCEL_FEE)) {
      long marketCancelFee = chainParamsConfig.getLong(
          ChainParameters.MARKET_CANCEL_FEE);
      if (marketCancelFee < 0 || marketCancelFee > 10_000_000_000L) {
        throw new ContractValidateException(
            "Bad MARKET_CANCEL_FEE parameter value, valid range is [0,10_000_000_000L]");
      }
      dynamicPropertiesStore.put("MARKET_CANCEL_FEE".getBytes(),
          ByteArray.fromLong(marketCancelFee));
      spec.commandLine().getOut().format("The MARKET_CANCEL_FEE has been modified as %d.",
          marketCancelFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_PBFT)) {
      long allowPbft = chainParamsConfig.getLong(
          ChainParameters.ALLOW_PBFT);
      checkLongBoolean(allowPbft);
      dynamicPropertiesStore.put("ALLOW_PBFT".getBytes(),
          ByteArray.fromLong(allowPbft));
      spec.commandLine().getOut().format("The ALLOW_PBFT has been modified as %d.",
          allowPbft).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TRANSACTION_FEE_POOL)) {
      long allowTransactionFeePool = chainParamsConfig.getLong(
          ChainParameters.ALLOW_TRANSACTION_FEE_POOL);
      checkLongBoolean(allowTransactionFeePool);
      dynamicPropertiesStore.put("ALLOW_TRANSACTION_FEE_POOL".getBytes(),
          ByteArray.fromLong(allowTransactionFeePool));
      spec.commandLine().getOut().format("The ALLOW_TRANSACTION_FEE_POOL has been modified as %d.",
          allowTransactionFeePool).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MAX_FEE_LIMIT)) {
      long maxFeeLimit = chainParamsConfig.getLong(
          ChainParameters.MAX_FEE_LIMIT);
      if (maxFeeLimit < 0 || maxFeeLimit > 10_000_000_000L) {
        throw new ContractValidateException(
            "Bad MAX_FEE_LIMIT parameter value, valid range is [0,10_000_000_000L]");
      }
      dynamicPropertiesStore.put("MAX_FEE_LIMIT".getBytes(),
          ByteArray.fromLong(maxFeeLimit));
      spec.commandLine().getOut().format("The MAX_FEE_LIMIT has been modified as %d.",
          maxFeeLimit).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_OPTIMIZE_BLACK_HOLE)) {
      long allowOptimizeBlackHole = chainParamsConfig.getLong(
          ChainParameters.ALLOW_OPTIMIZE_BLACK_HOLE);
      checkLongBoolean(allowOptimizeBlackHole);
      dynamicPropertiesStore.put("ALLOW_OPTIMIZE_BLACK_HOLE".getBytes(),
          ByteArray.fromLong(allowOptimizeBlackHole));
      spec.commandLine().getOut().format("The ALLOW_OPTIMIZE_BLACK_HOLE has been modified as %d.",
          allowOptimizeBlackHole).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_NEW_RESOURCE_MODEL)) {
      long allowNewResourceModel = chainParamsConfig.getLong(
          ChainParameters.ALLOW_NEW_RESOURCE_MODEL);
      checkLongBoolean(allowNewResourceModel);
      dynamicPropertiesStore.put("ALLOW_NEW_RESOURCE_MODEL".getBytes(),
          ByteArray.fromLong(allowNewResourceModel));
      spec.commandLine().getOut().format("The ALLOW_NEW_RESOURCE_MODEL has been modified as %d.",
          allowNewResourceModel).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_FREEZE)) {
      long allowTVMFreeze = chainParamsConfig.getLong(
          ChainParameters.ALLOW_TVM_FREEZE);
      checkLongBoolean(allowTVMFreeze);
      dynamicPropertiesStore.put("ALLOW_TVM_FREEZE".getBytes(),
          ByteArray.fromLong(allowTVMFreeze));
      spec.commandLine().getOut().format("The ALLOW_TVM_FREEZE has been modified as %d.",
          allowTVMFreeze).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_VOTE)) {
      long allowTVMVote = chainParamsConfig.getLong(
          ChainParameters.ALLOW_TVM_VOTE);
      checkLongBoolean(allowTVMVote);
      dynamicPropertiesStore.put("ALLOW_TVM_VOTE".getBytes(),
          ByteArray.fromLong(allowTVMVote));
      spec.commandLine().getOut().format("The ALLOW_TVM_VOTE has been modified as %d.",
          allowTVMVote).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_LONDON)) {
      long allowTVMLondon = chainParamsConfig.getLong(ChainParameters.ALLOW_TVM_LONDON);
      checkLongBoolean(allowTVMLondon);
      dynamicPropertiesStore.put("ALLOW_TVM_LONDON".getBytes(), ByteArray.fromLong(allowTVMLondon));
      spec.commandLine().getOut().format("The ALLOW_TVM_LONDON has been modified as %d.",
          allowTVMLondon).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_COMPATIBLE_EVM)) {
      long allowTVMCompatibleEVM = chainParamsConfig.getLong(ChainParameters.ALLOW_TVM_COMPATIBLE_EVM);
      checkLongBoolean(allowTVMCompatibleEVM);
      dynamicPropertiesStore.put("ALLOW_TVM_COMPATIBLE_EVM".getBytes(),
          ByteArray.fromLong(allowTVMCompatibleEVM));
      spec.commandLine().getOut().format("The ALLOW_TVM_COMPATIBLE_EVM has been modified as %d.",
          allowTVMCompatibleEVM).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_ACCOUNT_ASSET_OPTIMIZATION)) {
      long allowAccountAssetOptimization = chainParamsConfig.getLong(
          ChainParameters.ALLOW_ACCOUNT_ASSET_OPTIMIZATION);
      checkLongBoolean(allowAccountAssetOptimization);
      dynamicPropertiesStore.put("ALLOW_ACCOUNT_ASSET_OPTIMIZATION".getBytes(),
          ByteArray.fromLong(allowAccountAssetOptimization));
      spec.commandLine().getOut()
          .format("The ALLOW_ACCOUNT_ASSET_OPTIMIZATION has been modified as %d.",
              allowAccountAssetOptimization).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.FREE_NET_LIMIT)) {
      long freeNetLimit = chainParamsConfig.getLong(ChainParameters.FREE_NET_LIMIT);
      if (freeNetLimit < 0 || freeNetLimit > 100_000L) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [0,100_000]");
      }
      dynamicPropertiesStore.put("FREE_NET_LIMIT".getBytes(),
          ByteArray.fromLong(freeNetLimit));
      spec.commandLine().getOut().format("The FREE_NET_LIMIT has been modified as %d.",
          freeNetLimit).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.TOTAL_NET_LIMIT)) {
      long totalNetLimit = chainParamsConfig.getLong(ChainParameters.TOTAL_NET_LIMIT);
      if (totalNetLimit < 0 || totalNetLimit > 1_000_000_000_000L) {
        throw new ContractValidateException(
            "Bad chain parameter value, valid range is [0, 1_000_000_000_000L]");
      }
      dynamicPropertiesStore.put("TOTAL_NET_LIMIT".getBytes(),
          ByteArray.fromLong(totalNetLimit));
      spec.commandLine().getOut().format("The TOTAL_NET_LIMIT has been modified as %d.",
          totalNetLimit).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_HIGHER_LIMIT_FOR_MAX_CPU_TIME_OF_ONE_TX)) {
      long allowHigherLimitForMaxCpuTimeOfOneTx = chainParamsConfig.getLong(
          ChainParameters.ALLOW_HIGHER_LIMIT_FOR_MAX_CPU_TIME_OF_ONE_TX);
      checkLongBoolean(allowHigherLimitForMaxCpuTimeOfOneTx);
      dynamicPropertiesStore.put("ALLOW_HIGHER_LIMIT_FOR_MAX_CPU_TIME_OF_ONE_TX".getBytes(),
          ByteArray.fromLong(allowHigherLimitForMaxCpuTimeOfOneTx));
      spec.commandLine().getOut()
          .format("The ALLOW_HIGHER_LIMIT_FOR_MAX_CPU_TIME_OF_ONE_TX has been modified as %d.",
              allowHigherLimitForMaxCpuTimeOfOneTx).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_ASSET_OPTIMIZATION)) {
      long allowAssetOptimization = chainParamsConfig.getLong(
          ChainParameters.ALLOW_ASSET_OPTIMIZATION);
      checkLongBoolean(allowAssetOptimization);
      dynamicPropertiesStore.put("ALLOW_ASSET_OPTIMIZATION".getBytes(),
          ByteArray.fromLong(allowAssetOptimization));
      spec.commandLine().getOut().format("The ALLOW_ASSET_OPTIMIZATION has been modified as %d.",
              allowAssetOptimization).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_NEW_REWARD)) {
      long allowNewReward = chainParamsConfig.getLong(ChainParameters.ALLOW_NEW_REWARD);
      checkLongBoolean(allowNewReward);
      dynamicPropertiesStore.put("ALLOW_UPDATE_ACCOUNT_NAME".getBytes(),
          ByteArray.fromLong(allowNewReward));
      spec.commandLine().getOut().format("The ALLOW_NEW_REWARD has been modified as %d.",
              allowNewReward).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MEMO_FEE)) {
      long memoFee = chainParamsConfig.getLong(ChainParameters.MEMO_FEE);
      if (memoFee < 0 || memoFee > 1_000_000_000) {
        throw new ContractValidateException(
            "This value[MEMO_FEE] is only allowed to be in the range 0-1000_000_000");
      }
      dynamicPropertiesStore.put("MEMO_FEE".getBytes(), ByteArray.fromLong(memoFee));
      spec.commandLine().getOut().format("The MEMO_FEE has been modified as %d.",
              memoFee).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_DELEGATE_OPTIMIZATION)) {
      long allowDelegateOptimization = chainParamsConfig.getLong(
          ChainParameters.ALLOW_DELEGATE_OPTIMIZATION);
      checkLongBoolean(allowDelegateOptimization);
      dynamicPropertiesStore.put("ALLOW_DELEGATE_OPTIMIZATION".getBytes(),
          ByteArray.fromLong(allowDelegateOptimization));
      spec.commandLine().getOut().format("The ALLOW_DELEGATE_OPTIMIZATION has been modified as %d.",
              allowDelegateOptimization).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.UNFREEZE_DELAY_DAYS)) {
      long unfreezeDelayDays = chainParamsConfig.getLong(ChainParameters.UNFREEZE_DELAY_DAYS);
      if (unfreezeDelayDays < 1 || unfreezeDelayDays > 365) {
        throw new ContractValidateException(
            "This value[UNFREEZE_DELAY_DAYS] is only allowed to be in the range 1-365");
      }      dynamicPropertiesStore.put("UNFREEZE_DELAY_DAYS".getBytes(),
          ByteArray.fromLong(unfreezeDelayDays));
      spec.commandLine().getOut().format("The UNFREEZE_DELAY_DAYS has been modified as %d.",
              unfreezeDelayDays).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_OPTIMIZED_RETURN_VALUE_OF_CHAIN_ID)) {
      long value = chainParamsConfig.getLong(
          ChainParameters.ALLOW_OPTIMIZED_RETURN_VALUE_OF_CHAIN_ID);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_UPDATE_ACCOUNT_NAME".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut()
          .format("The ALLOW_OPTIMIZED_RETURN_VALUE_OF_CHAIN_ID has been modified as %d.",
              value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_DYNAMIC_ENERGY)) {
      long value = chainParamsConfig.getLong(ChainParameters.ALLOW_DYNAMIC_ENERGY);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_DYNAMIC_ENERGY".getBytes(), ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The ALLOW_DYNAMIC_ENERGY has been modified as %d.",
              value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.DYNAMIC_ENERGY_THRESHOLD)) {
      long value = chainParamsConfig.getLong(ChainParameters.DYNAMIC_ENERGY_THRESHOLD);
      checkLongRange(value);
      dynamicPropertiesStore.put("DYNAMIC_ENERGY_THRESHOLD".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The DYNAMIC_ENERGY_THRESHOLD has been modified as %d.",
              value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.DYNAMIC_ENERGY_INCREASE_FACTOR)) {
      long value = chainParamsConfig.getLong(ChainParameters.DYNAMIC_ENERGY_INCREASE_FACTOR);
      if (value < 0 || value > 10_000L) {
        throw new ContractValidateException(
            "This value[DYNAMIC_ENERGY_INCREASE_FACTOR] "
                + "is only allowed to be in the range 0-"
                + DYNAMIC_ENERGY_INCREASE_FACTOR_RANGE
        );
      }
      dynamicPropertiesStore.put("DYNAMIC_ENERGY_INCREASE_FACTOR".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut()
          .format("The DYNAMIC_ENERGY_INCREASE_FACTOR has been modified as %d.",
              value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_SHANGHAI)) {
      long value = chainParamsConfig.getLong(ChainParameters.ALLOW_TVM_SHANGHAI);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_TVM_SHANGHAI".getBytes(), ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The ALLOW_TVM_SHANGHAI has been modified as %d.",
          value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_CANCEL_ALL_UNFREEZE_V2)) {
      long value = chainParamsConfig.getLong(ChainParameters.ALLOW_CANCEL_ALL_UNFREEZE_V2);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_CANCEL_ALL_UNFREEZE_V2".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The ALLOW_CANCEL_ALL_UNFREEZE_V2 has been modified as %d.",
          value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.MAX_DELEGATE_LOCK_PERIOD)) {
      long value = chainParamsConfig.getLong(ChainParameters.MAX_DELEGATE_LOCK_PERIOD);
      long maxDelegateLockPeriod = ByteArray.toLong(
          dynamicPropertiesStore.get("MAX_DELEGATE_LOCK_PERIOD".getBytes()));
      if (value <= maxDelegateLockPeriod || value > ONE_YEAR_BLOCK_NUMBERS) {
        throw new ContractValidateException(
            "This value[MAX_DELEGATE_LOCK_PERIOD] is only allowed to be greater than "
                + maxDelegateLockPeriod + " and less than or equal to " + ONE_YEAR_BLOCK_NUMBERS
                + " !");
      }
      dynamicPropertiesStore.put("ALLOW_TVM_LIGHT_NODE".getBytes(), ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The MAX_DELEGATE_LOCK_PERIOD has been modified as %d.",
          value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_ENERGY_ADJUSTMENT)) {
      long value = chainParamsConfig.getLong(ChainParameters.ALLOW_ENERGY_ADJUSTMENT);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_ENERGY_ADJUSTMENT".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut()
          .format("The ALLOW_ENERGY_ADJUSTMENT has been modified as %d.", value);
    }

    if (chainParamsConfig.hasPath(ChainParameters.MAX_CREATE_ACCOUNT_TX_SIZE)) {
      long value = chainParamsConfig.getLong(ChainParameters.MAX_CREATE_ACCOUNT_TX_SIZE);
      if (value < CREATE_ACCOUNT_TRANSACTION_MIN_BYTE_SIZE
          || value > CREATE_ACCOUNT_TRANSACTION_MAX_BYTE_SIZE) {
        throw new ContractValidateException(
            "This value[MAX_CREATE_ACCOUNT_TX_SIZE] is only allowed to be greater than or equal "
                + "to " + CREATE_ACCOUNT_TRANSACTION_MIN_BYTE_SIZE + " and less than or equal to "
                + CREATE_ACCOUNT_TRANSACTION_MAX_BYTE_SIZE + "!");
      }
      dynamicPropertiesStore.put("MAX_CREATE_ACCOUNT_TX_SIZE".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut()
          .format("The MAX_CREATE_ACCOUNT_TX_SIZE has been modified as %d.", value);
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_STRICT_MATH)) {
      long value = chainParamsConfig.getLong(ChainParameters.ALLOW_STRICT_MATH);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_STRICT_MATH".getBytes(), ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The ALLOW_STRICT_MATH has been modified as %d.",
          value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.CONSENSUS_LOGIC_OPTIMIZATION)) {
      long value = chainParamsConfig.getLong(ChainParameters.CONSENSUS_LOGIC_OPTIMIZATION);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("CONSENSUS_LOGIC_OPTIMIZATION".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The CONSENSUS_LOGIC_OPTIMIZATION has been modified as %d.",
          value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_CANCUN)) {
      long value = chainParamsConfig.getLong(ChainParameters.ALLOW_TVM_CANCUN);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_TVM_CANCUN".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The ALLOW_TVM_CANCUN has been modified as %d.",
          value).println();
    }

    if (chainParamsConfig.hasPath(ChainParameters.ALLOW_TVM_BLOB)) {
      long value = chainParamsConfig.getLong(ChainParameters.ALLOW_TVM_BLOB);
      checkLongBoolean(value);
      dynamicPropertiesStore.put("ALLOW_TVM_BLOB".getBytes(),
          ByteArray.fromLong(value));
      spec.commandLine().getOut().format("The ALLOW_TVM_BLOB has been modified as %d.",
          value).println();
    }
  }

  private void checkLongRange(long value) throws ContractValidateException {
    if (value < 0 || value > ChainParameters.LONG_VALUE) {
      throw new ContractValidateException(
          "Bad chain parameter value, valid range is [0," + ChainParameters.LONG_VALUE + "]");
    }
  }

  private void checkLongBoolean(long value) throws ContractValidateException {
    if (value != 0 && value != 1) {
      throw new ContractValidateException(
          "Bad chain parameter value, valid value is 0 or 1");
    }
  }

  private void checkLongNegative(long value) throws ContractValidateException {
    if (value != 0 && value != 1 && value != -1) {
      throw new ContractValidateException(
          "Bad chain parameter value, valid value is -1, 0, 1.");
    }
  }


  public static byte[] getActiveWitness(List<ByteString> witnesses) {
    byte[] ba = new byte[witnesses.size() * Constant.ADDRESS_BYTE_ARRAY_LENGTH];
    int i = 0;
    for (ByteString address : witnesses) {
      System.arraycopy(address.toByteArray(), 0,
          ba, i * Constant.ADDRESS_BYTE_ARRAY_LENGTH, Constant.ADDRESS_BYTE_ARRAY_LENGTH);
      i++;
    }
    return ba;
  }
}
