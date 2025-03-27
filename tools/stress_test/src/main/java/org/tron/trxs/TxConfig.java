package org.tron.trxs;

import com.google.common.collect.Range;
import com.typesafe.config.Config;
import com.typesafe.config.ConfigValue;
import java.util.LinkedHashMap;
import java.util.Map;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import lombok.extern.slf4j.Slf4j;
import org.bouncycastle.util.encoders.Hex;
import org.tron.trident.core.key.KeyPair;
import org.tron.trident.utils.Base58Check;

import static org.tron.utils.Constant.*;

@Slf4j(topic = "trxConfig")
@NoArgsConstructor
public class TxConfig {

  private static final TxConfig INSTANCE = new TxConfig();

//  @Setter
//  @Getter
//  private String getAddressUrl;
//
//  @Setter
//  @Getter
//  private long getAddressStartNumber = 0;
//
//  @Setter
//  @Getter
//  private long getAddressEndNumber = 0;

  @Setter
  @Getter
  private int addressTotal = 0;

  @Setter
  @Getter
  private String addressDbPath;

  @Setter
  @Getter
  private boolean generateTx = true;

  @Setter
  @Getter
  private int totalTxCnt;

  @Setter
  @Getter
  private int singleTaskCnt;

  @Setter
  @Getter
  private Map<TxType, Range<Integer>> rangeMap = new LinkedHashMap<>();

  @Setter
  @Getter
  private String updateRefUrl;

  @Setter
  @Getter
  private String privateKey;

  @Setter
  @Getter
  private String fromAddress;

  @Setter
  @Getter
  private String addressListFile;

  @Setter
  @Getter
  private int trc10Id = 1000001;

  @Setter
  @Getter
  private String trc20ContractAddress = Hex
      .toHexString(Base58Check.base58ToBytes("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"));

  @Setter
  @Getter
  private boolean isRelay = false;

  @Setter
  @Getter
  private String relayUrl;

  @Setter
  @Getter
  private long relayStartNumber = 0;

  @Setter
  @Getter
  private long relayEndNumber = 0;

  @Setter
  @Getter
  private boolean broadcastGenerate = true;

  @Setter
  @Getter
  private boolean broadcastRelay = false;

  @Setter
  @Getter
  private int broadcastTpsLimit = 1000;

  @Setter
  @Getter
  private boolean saveTrxId = true;

  @Setter
  @Getter
  private String statisticUrl;

  @Setter
  @Getter
  private long statisticStartNumber = 0;

  @Setter
  @Getter
  private long statisticEndNumber = 0;


  public static void initParams(Config config) {
//    if (config.hasPath(GET_ADDRESS_URL)) {
//      INSTANCE.setGetAddressUrl(config.getString(GET_ADDRESS_URL));
//    }
//
//    if (config.hasPath(GET_ADDRESS_START_NUMBER)) {
//      INSTANCE.setGetAddressStartNumber(config.getLong(GET_ADDRESS_START_NUMBER));
//    }
//
//    if (config.hasPath(GET_ADDRESS_END_NUMBER)) {
//      INSTANCE.setGetAddressEndNumber(config.getLong(GET_ADDRESS_END_NUMBER));
//    }

    if (config.hasPath(ADDRESS_TOTAL)) {
      INSTANCE.setAddressTotal(config.getInt(ADDRESS_TOTAL));
    }

    if (config.hasPath(ADDRESS_DB_PATH)) {
      INSTANCE.setAddressDbPath(config.getString(ADDRESS_DB_PATH));
    }

    if (config.hasPath(GENERATE_TX_ENABLE)) {
      INSTANCE.setGenerateTx(config.getBoolean(GENERATE_TX_ENABLE));
    }

    if (config.hasPath(TOTAL_TX_CNT)
        && config.getInt(TOTAL_TX_CNT) > 0) {
      INSTANCE.setTotalTxCnt(config.getInt(TOTAL_TX_CNT));
    }

    if (INSTANCE.getTotalTxCnt() > INSTANCE.getAddressTotal()) {
      logger.error("the tx count should be less than the address total number");
      throw new IllegalArgumentException(
          "the tx count should be less than the address total number");
    }

    if (config.hasPath(SINGLE_TASK_TRX_COUNT)
        && config.getInt(SINGLE_TASK_TRX_COUNT) > 0) {
      INSTANCE.setSingleTaskCnt(config.getInt(SINGLE_TASK_TRX_COUNT));
    }

    if (config.hasPath(GENERATE_TX_TYPE)) {
      INSTANCE.setRangeMap(calculateRanges(config.getConfig(GENERATE_TX_TYPE)));
    }

    if (config.hasPath(UPDATE_REF_URL)) {
      INSTANCE.setUpdateRefUrl(config.getString(UPDATE_REF_URL));
    }

    if (config.hasPath(PRIVATE_KEY)) {
      if (config.getString(PRIVATE_KEY).length() != 64) {
        logger.error("private key is not valid.");
        throw new IllegalArgumentException("private key is not valid.");
      }

      INSTANCE.setPrivateKey(config.getString(PRIVATE_KEY));
      KeyPair keyPair = new KeyPair(INSTANCE.getPrivateKey());
      INSTANCE.setFromAddress(keyPair.toHexAddress());
    }

    if (config.hasPath(ADDRESS_LIST)) {
      INSTANCE.setAddressListFile(config.getString(ADDRESS_LIST));
    }

    if (config.hasPath(TRC10_ID)) {
      INSTANCE.setTrc10Id(config.getInt(TRC10_ID));
    }

    if (config.hasPath(TRC20_ADDRESS)) {
      INSTANCE.setTrc20ContractAddress(
          Hex.toHexString(Base58Check.base58ToBytes(config.getString(TRC20_ADDRESS))));
    }


    if (config.hasPath(RELAY_ENABLE)) {
      INSTANCE.setRelay(config.getBoolean(RELAY_ENABLE));
    }

    if (config.hasPath(RELAY_URL)) {
      INSTANCE.setRelayUrl(config.getString(RELAY_URL));
    }

    if (config.hasPath(RELAY_START_NUMBER)) {
      INSTANCE.setRelayStartNumber(config.getLong(RELAY_START_NUMBER));
    }

    if (config.hasPath(RELAY_END_NUMBER)) {
      INSTANCE.setRelayEndNumber(config.getLong(RELAY_END_NUMBER));
    }


    if (config.hasPath(BROADCAST_GENERATE)) {
      INSTANCE.setBroadcastGenerate(config.getBoolean(BROADCAST_GENERATE));
    }

    if (config.hasPath(BROADCAST_RELAY)) {
      INSTANCE.setBroadcastRelay(config.getBoolean(BROADCAST_RELAY));
    }

    if (config.hasPath(BROADCAST_TPS) && config.getInt(BROADCAST_TPS) > 0) {
      INSTANCE.setBroadcastTpsLimit(
          config.getInt(BROADCAST_TPS));
    }

    if (config.hasPath(SAVE_TRX_ID)) {
      INSTANCE.setSaveTrxId(config.getBoolean(SAVE_TRX_ID));
    }


    if (config.hasPath(STATISTIC_URL)) {
      INSTANCE.setStatisticUrl(config.getString(STATISTIC_URL));
    }

    if (config.hasPath(STATISTIC_START_NUMBER)) {
      INSTANCE.setStatisticStartNumber(config.getLong(STATISTIC_START_NUMBER));
    }

    if (config.hasPath(STATISTIC_END_NUMBER)) {
      INSTANCE.setStatisticEndNumber(config.getLong(STATISTIC_END_NUMBER));
    }
  }

  public static Map<TxType, Range<Integer>> calculateRanges(Config trxType) {
    Map<TxType, Range<Integer>> rangeMap = new LinkedHashMap<>();
    int start = 0;
    for (Map.Entry<String, ConfigValue> entry : trxType.entrySet()) {
      int value = trxType.getInt(entry.getKey());
      rangeMap
          .put(TxType.fromString(entry.getKey()), Range.closedOpen(start, start + value));
      start += value;
    }
    if (start != 100) {
      logger.error("transaction type sum not equals 100.");
      throw new IllegalArgumentException("transaction type sum not equals 100.");
    }

    return rangeMap;
  }

  public TxType findTransactionType(int number) {
    for (Map.Entry<TxType, Range<Integer>> entry : rangeMap.entrySet()) {
      if (entry.getValue().contains(number)) {
        return entry.getKey();
      }
    }
    return null;
  }

  public static TxConfig getInstance() {
    return INSTANCE;
  }

}
