package org.tron.trxs;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;
import lombok.extern.slf4j.Slf4j;
import org.tron.trident.core.ApiWrapper;
import org.tron.trident.core.exceptions.IllegalException;
import org.tron.trident.proto.Chain.Transaction;
import org.tron.trident.proto.Response.BlockExtention;
import org.tron.trident.proto.Response.BlockListExtention;
import org.tron.trident.proto.Response.TransactionExtention;

@Slf4j(topic = "relayTrxGenerator")
public class ReplayTxGenerator {

  private long startNum;
  private long endNum;
  private String outputFile;
  private String outputDir = "stress-test-output";

  private ApiWrapper apiWrapper;
  public static List<Transaction> transactionsOfReplay = new ArrayList<>();
  public static AtomicInteger indexOfReplayTransaction = new AtomicInteger();
  private int count = 0;

  private volatile boolean isReplayGenerate = true;
  private ConcurrentLinkedQueue<Transaction> transactions = new ConcurrentLinkedQueue<>();

  FileOutputStream fos = null;
  CountDownLatch countDownLatch = null;
  private ExecutorService savePool = Executors.newFixedThreadPool(1,
      r -> new Thread(r, "save-tx"));

  private ExecutorService generatePool = Executors.newFixedThreadPool(2,
      r -> new Thread(r, "relay-tx"));

  public ReplayTxGenerator(String outputFile) {
    File dir = new File(outputDir);
    if (!dir.exists()) {
      dir.mkdirs();
    }

    this.outputFile = outputDir + File.separator + outputFile;
    this.startNum = TxConfig.getInstance().getRelayStartNumber();
    this.endNum = TxConfig.getInstance().getRelayEndNumber();
    this.apiWrapper = new ApiWrapper(TxConfig.getInstance().getRelayUrl(),
        TxConfig.getInstance().getRelayUrl(),
        TxConfig.getInstance().getPrivateKey());
  }

  public ReplayTxGenerator() {
    this("relay-tx.csv");
  }

  private void consumerGenerateTransaction() throws IOException {
    if (transactions.isEmpty()) {
      try {
        Thread.sleep(100);
        return;
      } catch (InterruptedException e) {
        System.out.println(e);
      }
    }

    Transaction transaction = transactions.poll();
    transaction.writeDelimitedTo(fos);

    long count = countDownLatch.getCount();
    if (count % 1000 == 0) {
      fos.flush();
      logger.info(String.format("relay tx task, remain: %d, pending size: %d",
          countDownLatch.getCount(), transactions.size()));
    }

    countDownLatch.countDown();
  }

  public void start() {
    logger.info(
        String.format("extract the tx from block: %s to block: %s.", startNum, endNum));

    BlockListExtention blockList = null;
    Optional<List<BlockExtention>> result;
    int step = 50;
    long stepEndNumber;

    for (long i = startNum; i < endNum; i = i + step) {
      stepEndNumber = (i + step) > endNum ? endNum : i + step;
      try {
        blockList = apiWrapper.getBlockByLimitNext(i, stepEndNumber);
      } catch (IllegalException e) {
        logger.error("failed to get the blocks.");
        e.printStackTrace();
        System.exit(1);
      }
      result = Optional.ofNullable(blockList.getBlockList());
      if (result.isPresent()) {
        List<BlockExtention> blockExtentionList = result.get();
        if (blockExtentionList.size()> 0) {
          for (BlockExtention block : blockList.getBlockList()) {
            if (block.getTransactionsCount() > 0) {
              transactionsOfReplay
                  .addAll(block.getTransactionsList().stream().map(
                      TransactionExtention::getTransaction).collect(
                      Collectors.toList()));
            }
          }
        }
      }
      logger.info(String
          .format("extract the tx from block: %d to block: %d.", i, stepEndNumber));
    }

    logger.info("total relay transactions cnt: " + transactionsOfReplay.size());
    this.count = transactionsOfReplay.size();

    savePool.submit(() -> {
      while (isReplayGenerate) {
        try {
          consumerGenerateTransaction();
        } catch (IOException e) {
          e.printStackTrace();
        }
      }
    });

    try {
      fos = new FileOutputStream(this.outputFile);
      countDownLatch = new CountDownLatch(this.count);

      while (indexOfReplayTransaction.get() < transactionsOfReplay.size()) {
        transactions.add(transactionsOfReplay.get(indexOfReplayTransaction.get()));
        indexOfReplayTransaction.incrementAndGet();
      }

      countDownLatch.await();
      isReplayGenerate = false;

      fos.flush();
      fos.close();
    } catch (InterruptedException | IOException e) {
      e.printStackTrace();
    } finally {
      TxGenerator.shutDown(generatePool, savePool);
    }
  }
}
