package org.tron.trxs;

import static org.tron.trxs.BroadcastGenerate.getID;

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

@Slf4j(topic = "broadcastRelay")
public class BroadcastRelay {

  private volatile boolean isFinishSend = false;
  private final ConcurrentLinkedQueue<Transaction> transactionIDs = new ConcurrentLinkedQueue<>();

  private final TronNetService tronNetService;
  private final String output = "stress-test-output";

  private final ExecutorService saveTransactionIDPool = Executors
      .newFixedThreadPool(1, r -> new Thread(r, "save-relay-tx-id"));

  private final Random random = new Random(System.currentTimeMillis());

  public BroadcastRelay(TronNetService tronNetService) {
    this.tronNetService = tronNetService;
  }

  private void processTransactionID(int count, BufferedWriter bufferedWriter)
      throws InterruptedException, IOException {
    while (!isFinishSend || !transactionIDs.isEmpty()) {
      count++;
      if (transactionIDs.isEmpty()) {
        Thread.sleep(100);
      }

      Transaction transaction = transactionIDs.peek();
      Sha256Hash id = getID(transaction);
      bufferedWriter.write(id.toString());
      bufferedWriter.newLine();
      if (count % 1000 == 0) {
        bufferedWriter.flush();
        logger.info("transaction id size: {}", transactionIDs.size());
      }
      transactionIDs.poll();
    }
  }

  public void broadcastTransactions() {
    long trxCount = 0;
    boolean saveTrxId = TxConfig.getInstance().isSaveTrxId();
    if (saveTrxId) {
      saveTransactionIDPool.submit(() -> {
        int count = 0;
        try (
            FileWriter writer = new FileWriter(
                output + File.separator + "broadcast-relay-trxID.csv");
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

    long startTime = System.currentTimeMillis();
    logger.info("Start to process relay transaction broadcast task");
    try (FileInputStream fis = new FileInputStream(output + File.separator + "relay-tx.csv")) {
      Transaction transaction;
      int cnt = 0;
      long startTps = System.currentTimeMillis();
      long endTps;
      while ((transaction = Transaction.parseDelimitedFrom(fis)) != null) {
        trxCount++;
        if (cnt > TxConfig.getInstance().getBroadcastTpsLimit()) {
          endTps = System.currentTimeMillis();
          if (endTps - startTps < 1000) {
            Thread.sleep(1000 - (endTps - startTps));
          }
          cnt = 0;
          startTps = System.currentTimeMillis();
        } else {
          try {
            TransactionMessage message = new TransactionMessage(transaction);
            int peerCnt = tronNetService.fastBroadcastTransaction(message);
            while (peerCnt <= 0) {
              logger.warn("broadcast relay task has no available peers to broadcast, please wait");
              Thread.sleep(100);
              peerCnt = tronNetService.fastBroadcastTransaction(message);
            }
            if (trxCount % 1000 == 0) {
              logger.info("total broadcast tx num: {}", trxCount);
            }
          } catch (Exception e) {
            e.printStackTrace();
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
    } catch (IOException e) {
      e.printStackTrace();
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
    }

    long cost = System.currentTimeMillis() - startTime;
    logger.info("relay trx size: {}, cost: {}, tps: {}", trxCount, cost,
        1.0 * trxCount / cost * 1000);
    BroadcastGenerate.shutDown(saveTransactionIDPool);
  }
}
