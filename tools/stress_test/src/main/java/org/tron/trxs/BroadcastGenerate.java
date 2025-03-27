package org.tron.trxs;

import java.io.BufferedWriter;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileWriter;
import java.io.IOException;
import java.util.Random;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import lombok.extern.slf4j.Slf4j;
import org.tron.core.net.TronNetService;
import org.tron.core.net.message.adv.TransactionMessage;
import org.tron.protos.Protocol.Transaction;
import org.tron.trident.core.utils.Sha256Hash;

@Slf4j(topic = "broadcastGenerate")
public class BroadcastGenerate {

  private Integer dispatchCount;
  private String output = "stress-test-output";

  private volatile boolean isFinishSend = false;
  private ConcurrentLinkedQueue<Transaction> transactionIDs = new ConcurrentLinkedQueue<>();

  private TronNetService tronNetService;

  private static ExecutorService saveTransactionIDPool = Executors
      .newFixedThreadPool(1, r -> new Thread(r, "save-gen-tx-id"));

  private final Random random = new Random(System.currentTimeMillis());

  public BroadcastGenerate(TxConfig config, TronNetService tronNetService) {
    this.dispatchCount = config.getTotalTxCnt() / config.getSingleTaskCnt();
    this.tronNetService = tronNetService;
  }

  private void processTransactionID(int count, BufferedWriter bufferedWriter)
      throws InterruptedException, IOException {
    while (!isFinishSend || !transactionIDs.isEmpty()) {
      count++;
      if (transactionIDs.isEmpty()) {
        Thread.sleep(100);
        continue;
      }

      Transaction transaction = transactionIDs.peek();
      Sha256Hash id = getID(transaction);
      bufferedWriter.write(id.toString());
      bufferedWriter.newLine();
      if (count % 10000 == 0) {
        bufferedWriter.flush();
        logger.info("tx id size: {}", transactionIDs.size());
      }
      transactionIDs.poll();
    }
  }

  public void broadcastTransactions() throws IOException, InterruptedException {
    long trxCount = 0;
    boolean saveTrxId = TxConfig.getInstance().isSaveTrxId();
    int totalTask =
        TxConfig.getInstance().getTotalTxCnt() % TxConfig.getInstance().getSingleTaskCnt() == 0
            ? dispatchCount : dispatchCount + 1;
    long startTime = System.currentTimeMillis();
    for (int index = 0; index < totalTask; index++) {
      isFinishSend = false;
      if (saveTrxId) {
        int taskIndex = index;
        saveTransactionIDPool.submit(() -> {
          int count = 0;
          try (
              FileWriter writer = new FileWriter(
                  output + File.separator + "broadcast-generate-txID" + taskIndex + ".csv");
              BufferedWriter bufferedWriter = new BufferedWriter(writer)
          ) {
            processTransactionID(count, bufferedWriter);
          } catch (IOException e) {
            e.printStackTrace();
          } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
          }
        });
      }

      logger.info("Start to process broadcast generate trx task {}/{}", index + 1, totalTask);
      Transaction transaction;
      int cnt = 0;
      try (FileInputStream fis = new FileInputStream(
          output + File.separator + "generate-tx" + index + ".csv")) {
        long startTps = System.currentTimeMillis();
        long endTps;
        float currentTps;
        while ((transaction = Transaction.parseDelimitedFrom(fis)) != null) {
          trxCount++;
          if (cnt > TxConfig.getInstance().getBroadcastTpsLimit()) {
            endTps = System.currentTimeMillis();
            if (endTps - startTps <= 1000) {
              logger.info("real-time broadcast task {}/{} tps has reached: {}", index + 1, totalTask,
                  TxConfig.getInstance().getBroadcastTpsLimit());
              Thread.sleep(1000 - (endTps - startTps));
            } else {
              currentTps = cnt * 1000.0f / (endTps - startTps) ;
              logger.info("real-time broadcast task {}/{} tps is {}", index + 1, totalTask, currentTps);
            }
            cnt = 0;
            startTps = System.currentTimeMillis();
          } else {
            TransactionMessage message = new TransactionMessage(transaction);
            int peerCnt = tronNetService.fastBroadcastTransaction(message);
            while (peerCnt <= 0) {
              logger.warn("broadcast task {}/{} has no available peers to broadcast, please wait",
                  index + 1, totalTask);
              Thread.sleep(100);
              peerCnt = tronNetService.fastBroadcastTransaction(message);
            }
            if (saveTrxId) {
              transactionIDs.add(transaction);
            }
            cnt++;
          }
        }

        isFinishSend = true;
        while (saveTrxId && !transactionIDs.isEmpty()) {
          Thread.sleep(200);
        }
      } catch (Exception e) {
        e.printStackTrace();
        throw e;
      }
    }

    long cost = System.currentTimeMillis() - startTime;
    logger.info("broadcast generate tx size: {}, cost: {}, tps: {}",
        trxCount, cost, 1.0 * trxCount / cost * 1000);
    shutDown(saveTransactionIDPool);
  }

  static void shutDown(ExecutorService saveTransactionIDPool) {
    saveTransactionIDPool.shutdown();
    while (!saveTransactionIDPool.isTerminated()) {
      try {
        Thread.sleep(1000);
      } catch (InterruptedException e) {
        e.printStackTrace();
      }
    }
  }

  public static Sha256Hash getID(Transaction transaction) {
    return Sha256Hash.of(true, transaction.getRawData().toByteArray());
  }
}
