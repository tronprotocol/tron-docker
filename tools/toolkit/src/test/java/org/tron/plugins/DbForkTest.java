package org.tron.plugins;

import static org.tron.plugins.DbFork.getActiveWitness;
import static org.tron.plugins.utils.Constant.*;

import com.google.common.primitives.Bytes;
import com.google.common.primitives.Longs;
import com.google.protobuf.ByteString;
import com.typesafe.config.Config;
import com.typesafe.config.ConfigFactory;
import java.io.File;
import java.io.IOException;
import java.net.URL;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;
import org.junit.Assert;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.TemporaryFolder;
import org.rocksdb.RocksDBException;
import org.tron.common.utils.ByteArray;
import org.tron.common.utils.Commons;
import org.tron.core.capsule.AccountCapsule;
import org.tron.core.capsule.WitnessCapsule;
import org.tron.plugins.utils.Constant;
import org.tron.plugins.utils.FileUtils;
import org.tron.plugins.utils.db.DBInterface;
import org.tron.plugins.utils.db.DbTool;
import org.tron.protos.Protocol.Account;
import org.tron.protos.Protocol.Key;
import org.tron.protos.Protocol.Permission;
import org.tron.protos.Protocol.Permission.PermissionType;
import picocli.CommandLine;

public class DbForkTest {

  private DBInterface witnessStore;
  private DBInterface witnessScheduleStore;
  private DBInterface accountStore;
  private DBInterface dynamicPropertiesStore;
  private DBInterface accountAssetStore;
  private DBInterface assetIssueV2Store;
  private DBInterface contractStore;
  private DBInterface storageRowStore;

  @Rule
  public final TemporaryFolder folder = new TemporaryFolder();
  private String dbPath;
  private String forkPath;

  public void createDir() {
    String srcDir = dbPath + File.separator + "database";
    FileUtils.createDirIfNotExists(Paths.get(srcDir, WITNESS_STORE).toString());
    FileUtils.createDirIfNotExists(Paths.get(srcDir, WITNESS_SCHEDULE_STORE).toString());
    FileUtils.createDirIfNotExists(Paths.get(srcDir, ACCOUNT_STORE).toString());
    FileUtils.createDirIfNotExists(Paths.get(srcDir, DYNAMIC_PROPERTY_STORE).toString());

    FileUtils.createDirIfNotExists(Paths.get(srcDir, ACCOUNT_ASSET).toString());
    FileUtils.createDirIfNotExists(Paths.get(srcDir, ASSET_ISSUE_V2).toString());
    FileUtils.createDirIfNotExists(Paths.get(srcDir, CONTRACT_STORE).toString());
    FileUtils.createDirIfNotExists(Paths.get(srcDir, STORAGE_ROW_STORE).toString());

  }

  public void init() throws IOException, RocksDBException {

    String srcDir = dbPath + File.separator + "database";
    witnessStore = DbTool.getDB(srcDir, Constant.WITNESS_STORE);
    witnessScheduleStore = DbTool.getDB(srcDir, Constant.WITNESS_SCHEDULE_STORE);
    accountStore = DbTool.getDB(srcDir, Constant.ACCOUNT_STORE);
    dynamicPropertiesStore = DbTool.getDB(srcDir, Constant.DYNAMIC_PROPERTY_STORE);
    accountAssetStore = DbTool.getDB(srcDir, Constant.ACCOUNT_ASSET);
    assetIssueV2Store = DbTool.getDB(srcDir, Constant.ASSET_ISSUE_V2);
    contractStore = DbTool.getDB(srcDir, Constant.CONTRACT_STORE);
    storageRowStore = DbTool.getDB(srcDir, Constant.STORAGE_ROW_STORE);
  }

  public void close() {
    DbTool.close();
  }

  @Test
  public void testDbFork() throws IOException, RocksDBException {
    dbPath = folder.newFolder().toString();
    forkPath = getConfig("fork.conf");
    createDir();

    String[] args = new String[]{"-d",
        dbPath, "-c",
        forkPath};
    CommandLine cli = new CommandLine(new DbFork());
    Assert.assertEquals(0, cli.execute(args));

    init();
    Config forkConfig;
    File file = Paths.get(forkPath).toFile();
    if (file.exists() && file.isFile()) {
      forkConfig = ConfigFactory.parseFile(Paths.get(forkPath).toFile());
    } else {
      throw new IOException("Fork config file [" + forkPath + "] not exist!");
    }

    if (forkConfig.hasPath(WITNESS_KEY)) {
      List<? extends Config> witnesses = forkConfig.getConfigList(WITNESS_KEY);
      if (witnesses.isEmpty()) {
        System.out.println("no witness listed in the config.");
      }
      witnesses = witnesses.stream()
          .filter(c -> c.hasPath(WITNESS_ADDRESS))
          .collect(Collectors.toList());
      if (witnesses.isEmpty()) {
        System.out.println("no witness listed in the config.");
      }

      List<ByteString> witnessAddresses = witnesses.stream().map(
          w -> {
            ByteString address = ByteString.copyFrom(
                Commons.decodeFromBase58Check(w.getString(WITNESS_ADDRESS)));
            return address;
          }
      ).collect(Collectors.toList());
      Assert.assertArrayEquals(getActiveWitness(witnessAddresses),
          witnessScheduleStore.get(ACTIVE_WITNESSES));

      witnesses.stream().forEach(
          w -> {
            WitnessCapsule witnessCapsule = new WitnessCapsule(witnessStore.get(
                Commons.decodeFromBase58Check(w.getString(WITNESS_ADDRESS))));
            AccountCapsule accountCapsule = new AccountCapsule(accountStore.get(
                Commons.decodeFromBase58Check(w.getString(WITNESS_ADDRESS))));
            if (w.hasPath(WITNESS_VOTE)) {
              Assert.assertEquals(w.getLong(WITNESS_VOTE), witnessCapsule.getVoteCount());
            }
            if (w.hasPath(WITNESS_URL)) {
              Assert.assertEquals(w.getString(WITNESS_URL), witnessCapsule.getUrl());
            }
            Assert.assertTrue(accountCapsule.getIsWitness());
          }
      );
    }

    if (forkConfig.hasPath(ACCOUNTS_KEY)) {
      List<? extends Config> accounts = forkConfig.getConfigList(ACCOUNTS_KEY);
      if (accounts.isEmpty()) {
        System.out.println("no account listed in the config.");
      }
      accounts = accounts.stream()
          .filter(c -> c.hasPath(ACCOUNT_ADDRESS))
          .collect(Collectors.toList());
      if (accounts.isEmpty()) {
        System.out.println("no account listed in the config.");
      }
      accounts.stream().forEach(
          a -> {
            byte[] address = Commons.decodeFromBase58Check(a.getString(ACCOUNT_ADDRESS));
            AccountCapsule account = new AccountCapsule(accountStore.get(address));
            Assert.assertNotNull(account);
            if (a.hasPath(ACCOUNT_BALANCE)) {
              Assert.assertEquals(a.getLong(ACCOUNT_BALANCE), account.getBalance());
            }
            if (a.hasPath(ACCOUNT_NAME)) {
              Assert.assertArrayEquals(ByteArray.fromString(a.getString(ACCOUNT_NAME)),
                  account.getAccountName().toByteArray());
            }
            if (a.hasPath(ACCOUNT_TYPE)) {
              Assert.assertEquals(a.getString(ACCOUNT_TYPE), account.getType().toString());
            }
            if (a.hasPath(ACCOUNT_OWNER)) {
              Assert.assertArrayEquals(Commons.decodeFromBase58Check(a.getString(ACCOUNT_OWNER)),
                  account.getPermissionById(0).getKeys(0).getAddress().toByteArray());
            }
            if (a.hasPath(ACCOUNT_TRC10_ID) && a.hasPath(ACCOUNT_TRC10_BALANCE)
                && a.getLong(ACCOUNT_TRC10_BALANCE) > 0) {
              String trc10Id = a.getString(ACCOUNT_TRC10_ID);
              if (assetIssueV2Store.get(ByteArray.fromString(trc10Id)) != null) {
                if (account.getAssetOptimized()) {
                  byte[] k = Bytes.concat(address, ByteArray.fromString(trc10Id));
                  byte[] value = accountAssetStore.get(k);
                  Assert.assertEquals(a.getLong(ACCOUNT_TRC10_BALANCE), ByteArray.toLong(value));
                } else {
                  long value = account.getAssetMapV2().get(trc10Id);
                  Assert.assertEquals(a.getLong(ACCOUNT_TRC10_BALANCE), value);
                }
              }
            }
            if (a.getString(ACCOUNT_ADDRESS).equals("TLLM21wteSPs4hKjbxgmH1L6poyMjeTbHm")) {
              Assert.assertFalse(account.getIsWitness());
            }
          });
    }

    if (forkConfig.hasPath(LATEST_BLOCK_TIMESTAMP)) {
      long latestBlockHeaderTimestamp = forkConfig.getLong(LATEST_BLOCK_TIMESTAMP);
      Assert.assertEquals(latestBlockHeaderTimestamp,
          ByteArray.toLong(dynamicPropertiesStore.get(LATEST_BLOCK_HEADER_TIMESTAMP)));
    }

    if (forkConfig.hasPath(MAINTENANCE_INTERVAL)) {
      long maintenanceTimeInterval = forkConfig.getLong(MAINTENANCE_INTERVAL);
      Assert.assertEquals(maintenanceTimeInterval,
          ByteArray.toLong(dynamicPropertiesStore.get(MAINTENANCE_TIME_INTERVAL)));
    }

    if (forkConfig.hasPath(NEXT_MAINTENANCE_TIME)) {
      long nextMaintenanceTime = forkConfig.getLong(NEXT_MAINTENANCE_TIME);
      Assert.assertEquals(nextMaintenanceTime,
          ByteArray.toLong(dynamicPropertiesStore.get(MAINTENANCE_TIME)));
    }
    close();
  }

  @Test
  public void testDbForkAddsWitnessPermissionWhenMultiSignEnabled()
      throws IOException, RocksDBException {
    dbPath = folder.newFolder().toString();
    forkPath = getConfig("fork.conf");
    createDir();
    init();
    dynamicPropertiesStore.put(ALLOW_MULTI_SIGN, Longs.toByteArray(1L));
    byte[] activeDefaultOperations = new byte[32];
    Arrays.fill(activeDefaultOperations, (byte) 0x7f);
    dynamicPropertiesStore.put(ACTIVE_DEFAULT_OPERATIONS, activeDefaultOperations);
    close();

    String[] args = new String[]{"-d",
        dbPath, "-c",
        forkPath};
    CommandLine cli = new CommandLine(new DbFork());
    Assert.assertEquals(0, cli.execute(args));

    init();
    Config forkConfig = ConfigFactory.parseFile(Paths.get(forkPath).toFile());
    List<? extends Config> witnesses = forkConfig.getConfigList(WITNESS_KEY).stream()
        .filter(c -> c.hasPath(WITNESS_ADDRESS))
        .collect(Collectors.toList());

    witnesses.forEach(w -> {
      byte[] address = Commons.decodeFromBase58Check(w.getString(WITNESS_ADDRESS));
      AccountCapsule account = new AccountCapsule(accountStore.get(address));
      Assert.assertTrue(account.getInstance().hasOwnerPermission());
      Assert.assertTrue(account.getInstance().hasWitnessPermission());
      Assert.assertEquals(1, account.getInstance().getActivePermissionCount());
      Assert.assertArrayEquals(address,
          account.getInstance().getOwnerPermission().getKeys(0).getAddress().toByteArray());
      Assert.assertArrayEquals(address, account.getWitnessPermissionAddress());
      Assert.assertArrayEquals(activeDefaultOperations,
          account.getInstance().getActivePermission(0).getOperations().toByteArray());
    });

    close();
  }

  @Test
  public void testDbForkKeepsExistingWitnessPermissionWhenMultiSignEnabled()
      throws IOException, RocksDBException {
    dbPath = folder.newFolder().toString();
    forkPath = getConfig("fork.conf");
    createDir();
    init();
    dynamicPropertiesStore.put(ALLOW_MULTI_SIGN, Longs.toByteArray(1L));
    byte[] activeDefaultOperations = new byte[32];
    Arrays.fill(activeDefaultOperations, (byte) 0x7f);
    dynamicPropertiesStore.put(ACTIVE_DEFAULT_OPERATIONS, activeDefaultOperations);

    byte[] witnessAddress = Commons.decodeFromBase58Check("TS1hu4ZCcwBFYpQqUGoWy1GWBzamqxiT5W");
    byte[] customWitnessAddress = Commons.decodeFromBase58Check("TLLM21wteSPs4hKjbxgmH1L6poyMjeTbHm");
    Permission ownerPermission = AccountCapsule.createDefaultOwnerPermission(
        ByteString.copyFrom(witnessAddress));
    Permission activePermission = Permission.newBuilder()
        .setType(PermissionType.Active)
        .setId(2)
        .setPermissionName("active")
        .setThreshold(1)
        .setParentId(0)
        .setOperations(ByteString.copyFrom(activeDefaultOperations))
        .addKeys(Key.newBuilder()
            .setAddress(ByteString.copyFrom(witnessAddress))
            .setWeight(1)
            .build())
        .build();
    Permission witnessPermission = AccountCapsule.createDefaultWitnessPermission(
        ByteString.copyFrom(customWitnessAddress));
    AccountCapsule existingAccount = new AccountCapsule(Account.newBuilder()
        .setAddress(ByteString.copyFrom(witnessAddress))
        .setOwnerPermission(ownerPermission)
        .addActivePermission(activePermission)
        .setWitnessPermission(witnessPermission)
        .build());
    accountStore.put(witnessAddress, existingAccount.getData());
    close();

    String[] args = new String[]{"-d",
        dbPath, "-c",
        forkPath};
    CommandLine cli = new CommandLine(new DbFork());
    Assert.assertEquals(0, cli.execute(args));

    init();
    AccountCapsule account = new AccountCapsule(accountStore.get(witnessAddress));
    Assert.assertTrue(account.getInstance().hasWitnessPermission());
    Assert.assertArrayEquals(customWitnessAddress, account.getWitnessPermissionAddress());
    close();
  }

  private String getConfig(String config) {
    URL path = DbForkTest.class.getClassLoader().getResource(config);
    return path == null ? null : path.getPath();
  }
}
