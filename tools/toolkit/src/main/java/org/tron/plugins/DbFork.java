package org.tron.plugins;

import static java.lang.System.arraycopy;
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
import static org.tron.plugins.utils.Constant.LATEST_BLOCK_TIMESTAMP;
import static org.tron.plugins.utils.Constant.MAINTENANCE_INTERVAL;
import static org.tron.plugins.utils.Constant.MAINTENANCE_TIME;
import static org.tron.plugins.utils.Constant.MAINTENANCE_TIME_INTERVAL;
import static org.tron.plugins.utils.Constant.MAX_ACTIVE_WITNESS_NUM;
import static org.tron.plugins.utils.Constant.NEXT_MAINTENANCE_TIME;
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

  private DBInterface witnessStore;
  private DBInterface witnessScheduleStore;
  private DBInterface accountStore;
  private DBInterface dynamicPropertiesStore;
  private DBInterface accountAssetStore;
  private DBInterface assetIssueV2Store;
  private DBInterface contractStore;
  private DBInterface storageRowStore;

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

  private void initStore() throws IOException, RocksDBException {
    String srcDir = database + File.separator + "database";
    witnessStore = DbTool.getDB(srcDir, WITNESS_STORE);
    witnessScheduleStore = DbTool.getDB(srcDir, WITNESS_SCHEDULE_STORE);
    accountStore = DbTool.getDB(srcDir, ACCOUNT_STORE);
    dynamicPropertiesStore = DbTool.getDB(srcDir, DYNAMIC_PROPERTY_STORE);
    accountAssetStore = DbTool.getDB(srcDir, ACCOUNT_ASSET);
    assetIssueV2Store = DbTool.getDB(srcDir, ASSET_ISSUE_V2);
    contractStore = DbTool.getDB(srcDir, CONTRACT_STORE);
    storageRowStore = DbTool.getDB(srcDir, STORAGE_ROW_STORE);
  }

  @Override
  public Integer call() throws IOException, RocksDBException {
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
      forkConfig = ConfigFactory.parseFile(Paths.get(config).toFile());
    } else {
      logger.error("Fork config file [" + config + "] not exists!");
      spec.commandLine().getErr().format("Fork config file: %s not exists!", config).println();
      System.exit(-1);
    }

    initStore();

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

    if (forkConfig.hasPath(WITNESS_KEY)) {
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

    if (forkConfig.hasPath(ACCOUNTS_KEY)) {
      List<? extends Config> accounts = forkConfig.getConfigList(ACCOUNTS_KEY);
      if (accounts.isEmpty()) {
        spec.commandLine().getOut().println("no account listed in the config.");
      }

      accounts = accounts.stream()
          .filter(c -> c.hasPath(ACCOUNT_ADDRESS))
          .collect(Collectors.toList());

      if (accounts.isEmpty()) {
        spec.commandLine().getOut().println("no account listed in the config.");
      }

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

    if (forkConfig.hasPath(TRC20_CONTRACTS_KEY)) {
      List<? extends Config> trc20Contracts = forkConfig.getConfigList(TRC20_CONTRACTS_KEY);
      if (trc20Contracts.isEmpty()) {
        spec.commandLine().getOut().println("no TRC20 contract listed in the config.");
      }

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

    if (forkConfig.hasPath(LATEST_BLOCK_TIMESTAMP)
        && forkConfig.getLong(LATEST_BLOCK_TIMESTAMP) > 0) {
      long latestBlockHeaderTimestamp = forkConfig.getLong(LATEST_BLOCK_TIMESTAMP);
      dynamicPropertiesStore
          .put(LATEST_BLOCK_HEADER_TIMESTAMP, ByteArray.fromLong(latestBlockHeaderTimestamp));
      logger.info("The latest block header timestamp has been modified as {}.",
          latestBlockHeaderTimestamp);
      spec.commandLine().getOut().format("The latest block header timestamp has been modified "
          + "as %d.", latestBlockHeaderTimestamp).println();
    }

    if (forkConfig.hasPath(MAINTENANCE_INTERVAL)
        && forkConfig.getLong(MAINTENANCE_INTERVAL) > 0) {
      long maintenanceTimeInterval = forkConfig.getLong(MAINTENANCE_INTERVAL);
      dynamicPropertiesStore
          .put(MAINTENANCE_TIME_INTERVAL, ByteArray.fromLong(maintenanceTimeInterval));
      logger.info("The maintenance time interval has been modified as {}.",
          maintenanceTimeInterval);
      spec.commandLine().getOut().format("The maintenance time interval has been modified as %d.",
          maintenanceTimeInterval).println();
    }

    if (forkConfig.hasPath(NEXT_MAINTENANCE_TIME)
        && forkConfig.getLong(NEXT_MAINTENANCE_TIME) > 0) {
      long nextMaintenanceTime = forkConfig.getLong(NEXT_MAINTENANCE_TIME);
      dynamicPropertiesStore.put(MAINTENANCE_TIME, ByteArray.fromLong(nextMaintenanceTime));
      logger.info("The next maintenance time has been modified as {}.",
          nextMaintenanceTime);
      spec.commandLine().getOut().format("The next maintenance time has been modified as %d.",
          nextMaintenanceTime).println();
    }

    DbTool.close();
    return 0;
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
