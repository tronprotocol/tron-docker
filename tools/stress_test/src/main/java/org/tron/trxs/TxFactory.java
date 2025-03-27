package org.tron.trxs;

import com.google.protobuf.Any;
import com.google.protobuf.ByteString;
import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.util.List;
import java.util.Objects;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;
import lombok.Getter;
import lombok.Setter;
import lombok.extern.slf4j.Slf4j;
import org.bouncycastle.util.encoders.Hex;
import org.tron.trident.core.ApiWrapper;
import org.tron.trident.core.exceptions.IllegalException;
import org.tron.trident.core.key.KeyPair;
import org.tron.trident.core.transaction.BlockId;
import org.tron.trident.core.utils.ByteArray;
import org.tron.trident.core.utils.Sha256Hash;
import org.tron.trident.crypto.Hash;
import org.tron.trident.proto.Chain;
import org.tron.trident.proto.Chain.Block;
import org.tron.trident.proto.Chain.Transaction;
import org.tron.trident.proto.Chain.Transaction.Contract.ContractType;
import org.tron.trident.proto.Contract;
import org.tron.trident.proto.Response.TransactionExtention;
import org.tron.trident.utils.Base58Check;

@Slf4j(topic = "trxFactory")
public class TxFactory {

  private static final TxFactory INSTANCE = new TxFactory();

  private static final long feeLimit = 2000000000L;
  private static final long transferAmount = 1L;
  private static final long transferTrc10Amount = 1L;
  private static final long transferTrc20Amount = 1L;

  @Setter
  @Getter
  private long validPeriod = 24 * 60 * 60 * 1000L;

  @Setter
  @Getter
  private AtomicLong time = new AtomicLong(System.currentTimeMillis() + validPeriod);

  @Setter
  @Getter
  private ByteString refBlockNum;

  @Setter
  @Getter
  private ByteString refBlockHash;

  @Setter
  @Getter
  private KeyPair keyPair;

  private ConcurrentLinkedQueue<String> addressQueue = new ConcurrentLinkedQueue<>();

  @Setter
  @Getter
  private TxConfig config;

  private ScheduledExecutorService updateExecutor = Executors
      .newSingleThreadScheduledExecutor();

  @Getter
  private ApiWrapper apiWrapper;

  private String methodSign = "transfer(address,uint256)";

  public static void initInstance() {
    INSTANCE.config = TxConfig.getInstance();
    INSTANCE.keyPair = new KeyPair(INSTANCE.config.getPrivateKey());
    INSTANCE.loadAddressList(INSTANCE.config.getAddressListFile());

    long expirationTime = System.currentTimeMillis() + INSTANCE.validPeriod;
    INSTANCE.time.set(expirationTime);
    INSTANCE.apiWrapper = new ApiWrapper(INSTANCE.config.getUpdateRefUrl(),
        INSTANCE.config.getUpdateRefUrl(), INSTANCE.config.getPrivateKey());
    INSTANCE.update();
    INSTANCE.updateTrxReference();
  }

  public void updateTrxReference() {
    updateExecutor.scheduleWithFixedDelay(() -> {
      try {
        update();
      } catch (Exception e) {
        logger.error("failed to update the transaction reference");
        e.printStackTrace();
        System.exit(1);
      }
    }, 60, 60, TimeUnit.SECONDS);

  }

  private void update() {
    logger.info("begin to update the tx reference");
    time.set(Math.max(System.currentTimeMillis() + validPeriod, time.get()));
    Block block = null;
    try {
      block = apiWrapper.getNowBlock();
    } catch (IllegalException e) {
      logger.error("failed to get the block");
      e.printStackTrace();
      System.exit(1);
    }
    long blockNum = block.getBlockHeader().getRawData().getNumber() - 1;
    byte[] blockHash = block.getBlockHeader().getRawData().getParentHash().toByteArray();
    refBlockHash = ByteString.copyFrom(ByteArray.subArray(blockHash, 8, 16));
    refBlockNum = ByteString.copyFrom(ByteArray.subArray(ByteArray.fromLong(blockNum), 6, 8));

    BlockId blockId = new BlockId(blockHash, blockNum);
    long expiration = time.incrementAndGet();
    apiWrapper.enableLocalCreate(blockId, expiration);

    logger.info("finish updating the tx reference");
  }

  public void loadAddressList(String filePath) {
    try (BufferedReader reader = new BufferedReader(
        new FileReader(filePath))) {
      String line;
      while ((line = reader.readLine()) != null && line.length() == 34) {
        addressQueue.offer(line);
      }
      logger.info("load address success, size: {}", addressQueue.size());
    } catch (IOException e) {
      e.printStackTrace();
    }
  }

  public void close() {
    updateExecutor.shutdown();
  }

  public Transaction getTransferTx() throws IllegalException {
    TransactionExtention transaction = apiWrapper
        .transfer(config.getFromAddress(), addressQueue.poll(),
            transferAmount);
    return apiWrapper.signTransaction(transaction);
  }

  public Transaction getTransferTrc10() throws IllegalException {
    TransactionExtention transaction = apiWrapper
        .transferTrc10(config.getFromAddress(), addressQueue.poll(),
            config.getTrc10Id(),
            transferTrc10Amount);
    return apiWrapper.signTransaction(transaction);
  }

  public Transaction getTransferTrc20() throws Exception {
    byte[] toAddress = Base58Check.base58ToBytes(Objects.requireNonNull(addressQueue.poll()));
    String contractData = createContractData(INSTANCE.methodSign, toAddress, transferTrc20Amount);
    TransactionExtention transaction = apiWrapper
        .triggerContract(config.getFromAddress(), config.getTrc20ContractAddress(), contractData,
            0L, 0L, null, feeLimit);
    return apiWrapper.signTransaction(transaction);
  }

  public Transaction createTransferTx() {
    byte[] toAddress = Base58Check.base58ToBytes(Objects.requireNonNull(addressQueue.poll()));
    Contract.TransferContract contract = Contract.TransferContract.newBuilder()
        .setOwnerAddress(ByteString.fromHex(config.getFromAddress()))
        .setToAddress(ByteString.copyFrom(toAddress))
        .setAmount(transferAmount)
        .build();

    Transaction transaction = createTransaction(contract, ContractType.TransferContract);
    transaction = sign(transaction);
    return transaction;
  }

  public Transaction createTransferTrc10() {
    byte[] toAddress = Base58Check.base58ToBytes(Objects.requireNonNull(addressQueue.poll()));
    Contract.TransferAssetContract contract = Contract.TransferAssetContract.newBuilder()
        .setAssetName(ByteString.copyFromUtf8(String.valueOf(config.getTrc10Id())))
        .setOwnerAddress(ByteString.fromHex(config.getFromAddress()))
        .setToAddress(ByteString.copyFrom(toAddress))
        .setAmount(transferTrc20Amount)
        .build();

    Transaction transaction = createTransaction(contract, ContractType.TransferAssetContract);
    transaction = sign(transaction);
    return transaction;
  }

  public Transaction createTransferTrc20() {
    byte[] toAddress = Base58Check.base58ToBytes(Objects.requireNonNull(addressQueue.poll()));
    String contractData = createContractData(INSTANCE.methodSign, toAddress, transferTrc20Amount);
    Contract.TriggerSmartContract contract = Contract.TriggerSmartContract.newBuilder()
        .setOwnerAddress(ByteString.fromHex(config.getFromAddress()))
        .setContractAddress(ByteString.fromHex(config.getTrc20ContractAddress()))
        .setData(ByteString.fromHex(contractData))
        .setCallValue(0L)
        .setTokenId(Long.parseLong("0"))
        .setCallTokenValue(0L)
        .build();

    Transaction transaction = createTransaction(contract, ContractType.TriggerSmartContract);
    Transaction.raw raw = transaction.getRawData().toBuilder().setFeeLimit(feeLimit).build();
    transaction = sign(transaction.toBuilder().setRawData(raw).build());
    return transaction;
  }

  private static String createContractData(String methodSign, byte[] receiver, long amount) {
    byte[] selector = new byte[4];
    System.arraycopy(Hash.sha3(methodSign.getBytes()), 0, selector, 0, 4);

    byte[] params = new byte[64];
    System.arraycopy(receiver, 1, params, 12, 20);
    System.arraycopy(ByteArray.fromLong(amount), 0, params, 56, 8);
    return Hex.toHexString(selector) + Hex.toHexString(params);
  }


  public Transaction createTransaction(com.google.protobuf.Message message,
      ContractType contractType) {
    Transaction.raw.Builder transactionBuilder = Transaction.raw.newBuilder().addContract(
        Transaction.Contract.newBuilder().setType(contractType).setParameter(
            Any.pack(message)).build());

    Transaction transaction = Transaction.newBuilder().setRawData(transactionBuilder.build())
        .build();
    long expiration = time.incrementAndGet();
    return setReferenceAndExpiration(transaction, expiration);
  }

  private Transaction setReference(Transaction transaction, long blockNum,
      byte[] blockHash) {
    byte[] refBlockNum = ByteArray.fromLong(blockNum);
    Transaction.raw rawData = transaction.getRawData().toBuilder()
        .setRefBlockHash(ByteString.copyFrom(blockHash))
        .setRefBlockBytes(ByteString.copyFrom(refBlockNum))
        .build();
    return transaction.toBuilder().setRawData(rawData).build();
  }

  private Transaction setReferenceAndExpiration(Transaction transaction, long expiration) {
    Transaction.raw rawData = transaction.getRawData().toBuilder()
        .setExpiration(expiration)
        .setRefBlockHash(refBlockHash)
        .setRefBlockBytes(refBlockNum)
        .build();
    return transaction.toBuilder().setRawData(rawData).build();
  }

  public static Transaction setExpiration(Transaction transaction, long expiration) {
    Transaction.raw rawData = transaction.getRawData().toBuilder().setExpiration(expiration)
        .build();
    return transaction.toBuilder().setRawData(rawData).build();
  }

  public Transaction sign(Transaction transaction) {
    Transaction.Builder transactionBuilderSigned = transaction.toBuilder();
    byte[] hash = Sha256Hash.hash(true, transaction.getRawData().toByteArray());
    List<Chain.Transaction.Contract> listContract = transaction.getRawData().getContractList();
    for (int i = 0; i < listContract.size(); i++) {
      byte[] signature = KeyPair.signTransaction(hash, keyPair);
      transactionBuilderSigned.addSignature(ByteString.copyFrom(signature));
    }

    transaction = transactionBuilderSigned.build();
    return transaction;
  }

  public static TxFactory getInstance() {
    return INSTANCE;
  }
}
