package org.tron.trxs;

import java.util.HashMap;
import java.util.Map;

public enum TxType {
  TRANSFER("transfer"),
  TRANSFER_TRC10("transferTrc10"),
  TRANSFER_TRC20("transferTrc20");

  private final String type;

  TxType(String type) {
    this.type = type;
  }

  private static final Map<String, TxType> stringToTypeMap = new HashMap<>();

  static {
    for (TxType type : TxType.values()) {
      stringToTypeMap.put(type.type, type);
    }
  }

  public static TxType fromString(String type) {
    return stringToTypeMap.get(type);
  }

  @Override
  public String toString() {
    return this.type;
  }

}
