package org.tron;

import com.google.common.primitives.Bytes;
import com.google.protobuf.ByteString;
import com.typesafe.config.Config;
import com.typesafe.config.ConfigFactory;
import java.io.File;
import java.io.IOException;
import java.net.URL;
import java.nio.file.Paths;
import java.util.List;
import java.util.stream.Collectors;
import org.junit.Assert;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.TemporaryFolder;
import org.tron.common.utils.ByteArray;
import org.tron.common.utils.Commons;
import org.tron.core.capsule.AccountCapsule;
import org.tron.core.capsule.WitnessCapsule;
import org.tron.db.TronDatabase;
import org.tron.utils.Utils;
import picocli.CommandLine;

import static org.tron.utils.Constant.*;

public class DBForkTest {

  private TronDatabase witnessStore;
  private TronDatabase witnessScheduleStore;
  private TronDatabase accountStore;
  private TronDatabase dynamicPropertiesStore;
  private TronDatabase accountAssetStore;
  private TronDatabase assetIssueV2Store;
  private TronDatabase contractStore;
  private TronDatabase storageRowStore;

  @Rule
  public final TemporaryFolder folder = new TemporaryFolder();
  private String dbPath;
  private String forkPath;
  private String dbEngine = "leveldb";

  public void init() {
    witnessStore = new TronDatabase(dbPath, WITNESS_STORE, dbEngine);
    witnessScheduleStore = new TronDatabase(dbPath, WITNESS_SCHEDULE_STORE, dbEngine);
    accountStore = new TronDatabase(dbPath, ACCOUNT_STORE, dbEngine);
    dynamicPropertiesStore = new TronDatabase(dbPath, DYNAMIC_PROPERTY_STORE, dbEngine);
    accountAssetStore = new TronDatabase(dbPath, ACCOUNT_ASSET, dbEngine);
    assetIssueV2Store = new TronDatabase(dbPath, ASSET_ISSUE_V2, dbEngine);
    contractStore = new TronDatabase(dbPath, CONTRACT_STORE, dbEngine);
    storageRowStore = new TronDatabase(dbPath, STORAGE_ROW_STORE, dbEngine);
  }

  public void close() {
    witnessStore.close();
    witnessScheduleStore.close();
    accountStore.close();
    dynamicPropertiesStore.close();
    accountAssetStore.close();
    assetIssueV2Store.close();
    contractStore.close();
    storageRowStore.close();
  }

  @Test
  public void testDbFork() throws IOException {
    dbPath = folder.newFolder().toString();
    forkPath = getConfig("fork.conf");

    String[] args = new String[]{"-d",
        dbPath, "-c",
        forkPath};
    CommandLine cli = new CommandLine(new DBFork());
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
      Assert.assertArrayEquals(Utils.getActiveWitness(witnessAddresses),
          witnessScheduleStore.get(ACTIVE_WITNESSES));

      witnesses.stream().forEach(
          w -> {
            WitnessCapsule witnessCapsule = new WitnessCapsule(witnessStore.get(
                Commons.decodeFromBase58Check(w.getString(WITNESS_ADDRESS))));
            if (w.hasPath(WITNESS_VOTE)) {
              Assert.assertEquals(w.getLong(WITNESS_VOTE), witnessCapsule.getVoteCount());
            }
            if (w.hasPath(WITNESS_URL)) {
              Assert.assertEquals(w.getString(WITNESS_URL), witnessCapsule.getUrl());
            }
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
              if (assetIssueV2Store.has(ByteArray.fromString(trc10Id))) {
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

  private static String getConfig(String config) {
    URL path = DBForkTest.class.getClassLoader().getResource(config);
    return path == null ? null : path.getPath();
  }
}
