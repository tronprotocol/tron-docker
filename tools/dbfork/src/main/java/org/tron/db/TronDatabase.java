package org.tron.db;

import java.nio.file.Paths;
import java.util.Objects;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import org.iq80.leveldb.WriteOptions;
import org.rocksdb.DirectComparator;
import org.tron.common.setting.RocksDbSettings;
import org.tron.common.utils.DbOptionalsUtils;
import org.tron.core.db.common.DbSourceInter;

@Slf4j(topic = "DB")
public class TronDatabase {

  protected DbSourceInter<byte[]> dbSource;
  @Getter
  private String dbName;

  public TronDatabase(String outputDirectory, String dbName, String dbEngine) {
    this.dbName = dbName;

    if ("LEVELDB".equalsIgnoreCase(dbEngine)) {
      dbSource =
          new LevelDbDataSourceImpl(outputDirectory,
              dbName,
              DbOptionalsUtils.createDefaultDbOptions(),
              new WriteOptions().sync(false));
    } else if ("ROCKSDB".equalsIgnoreCase(dbEngine)) {
      String parentName = Paths.get(outputDirectory,
          "database").toString();
      dbSource =
          new RocksDbDataSourceImpl(parentName, dbName, RocksDbSettings.getDefaultSettings(),
              getDirectComparator());
    } else {
      log.error("invalid db engine: {}", dbEngine);
      System.exit(-1);
    }

    dbSource.initDB();
  }

  protected DirectComparator getDirectComparator() {
    return null;
  }

  public void put(byte[] key, byte[] value) {
    if (Objects.isNull(key) || Objects.isNull(value)) {
      return;
    }

    dbSource.putData(key, value);
  }

  public byte[] get(byte[] key) {
    if (Objects.isNull(key)) {
      return null;
    }

    return dbSource.getData(key);
  }

  public void delete(byte[] key) {
    if (Objects.isNull(key)) {
      return;
    }

    dbSource.deleteData(key);
  }

  public boolean has(byte[] key) {
    return dbSource.getData(key) != null;
  }

  /**
   * reset the database.
   */
  public void reset() {
    dbSource.resetDb();
  }

  public void close() {
    log.info("******** Begin to close {}. ********", getName());
    try {
      dbSource.closeDB();
    } catch (Exception e) {
      log.warn("Failed to close {}.", getName(), e);
    } finally {
      log.info("******** End to close {}. ********", getName());
    }
  }

  public String getName() {
    return this.getClass().getSimpleName();
  }

}
