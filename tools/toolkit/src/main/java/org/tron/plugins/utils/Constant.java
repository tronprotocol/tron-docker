package org.tron.plugins.utils;

public class Constant {

  public static final String WITNESS_KEY = "witnesses";
  public static final String WITNESS_ADDRESS = "address";
  public static final String WITNESS_URL = "url";
  public static final String WITNESS_VOTE = "voteCount";
  public static final String ACCOUNTS_KEY = "accounts";
  public static final String ACCOUNT_NAME = "accountName";
  public static final String ACCOUNT_TYPE = "accountType";
  public static final String ACCOUNT_ADDRESS = "address";
  public static final String ACCOUNT_BALANCE = "balance";
  public static final String ACCOUNT_OWNER = "owner";
  public static final String ACCOUNT_TRC10_ID = "trc10Id";
  public static final String ACCOUNT_TRC10_BALANCE = "trc10Balance";

  public static final String TRC20_CONTRACTS_KEY = "trc20Contracts";
  public static final String TRC20_CONTRACT_ADDRESS = "contractAddress";
  public static final String TRC20_BALANCES_POSITION = "balancesSlotPosition";
  public static final String TRC20_ACCOUNT = "address";
  public static final String TRC20_BALANCE = "balance";

  public static final String LATEST_BLOCK_TIMESTAMP = "latestBlockHeaderTimestamp";
  public static final String MAINTENANCE_INTERVAL = "maintenanceTimeInterval";
  public static final String NEXT_MAINTENANCE_TIME = "nextMaintenanceTime";
  public static final int MAX_ACTIVE_WITNESS_NUM = 27;

  public static final String WITNESS_STORE = "witness";
  public static final String WITNESS_SCHEDULE_STORE = "witness_schedule";
  public static final String ACCOUNT_STORE = "account";
  public static final String DYNAMIC_PROPERTY_STORE = "properties";
  public static final String ASSET_ISSUE_V2 = "asset-issue-v2";
  public static final String ACCOUNT_ASSET = "account-asset";
  public static final String CONTRACT_STORE = "contract";
  public static final String STORAGE_ROW_STORE = "storage-row";

  public static final byte[] LATEST_BLOCK_HEADER_TIMESTAMP = "latest_block_header_timestamp"
      .getBytes();
  public static final byte[] MAINTENANCE_TIME_INTERVAL = "MAINTENANCE_TIME_INTERVAL".getBytes();
  public static final byte[] MAINTENANCE_TIME = "NEXT_MAINTENANCE_TIME".getBytes();
  public static final byte[] ACTIVE_WITNESSES = "active_witnesses".getBytes();
  public static final int ADDRESS_BYTE_ARRAY_LENGTH = 21;
}
