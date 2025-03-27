package org.tron.utils;

import java.io.BufferedWriter;
import java.io.FileWriter;
import java.io.IOException;
import lombok.Setter;
import lombok.extern.slf4j.Slf4j;
import org.tron.trident.core.ApiWrapper;
import org.tron.trident.core.exceptions.IllegalException;
import org.tron.trident.proto.Response.BlockExtention;

@Slf4j(topic = "statistic")
public class Statistic {

  @Setter
  private static ApiWrapper apiWrapper;

  @Setter
  private static int broadcastTpsLimit;

  @Setter
  private static int totalGenerateTxCnt;

  public static void result(long startBlock, long endBlock, String output) throws IllegalException {
    logger.info("TPS static range: start block: {}, end block: {}", startBlock, endBlock);

    BlockExtention block;
    long startNumber = 0, endNumber = 0;
    for (long i = startBlock; i < endBlock; i++) {
      block = apiWrapper.getBlockByNum(i);
      if (block.getTransactionsCount() > 0) {
        startNumber = i;
        break;
      }
    }

    for (long i = endBlock; i >= startBlock; i--) {
      block = apiWrapper.getBlockByNum(i);
      if (block.getTransactionsCount() > 0) {
        endNumber = i;
        break;
      }
    }

    if (startNumber < endNumber) {
      logger.info("startNumber: {}, endNumber: {}", startNumber, endNumber);
    } else {
      logger.error("invalid startNumber: {}, endNumber: {}", startNumber, endNumber);
      return;
    }

    int maxTrxCntInOneBlock = 0;
    int minTrxCntInOneBlock = 1000000;
    int trxCnt = 0;
    int totalTrxCnt = 0;

    for (long i = startNumber; i < endNumber; i++) {
      block = apiWrapper.getBlockByNum(i);
      trxCnt = block.getTransactionsCount();
      totalTrxCnt += trxCnt;

      if (trxCnt > maxTrxCntInOneBlock) {
        maxTrxCntInOneBlock = trxCnt;
      }
      if (trxCnt < minTrxCntInOneBlock) {
        minTrxCntInOneBlock = trxCnt;
      }
    }

    long expectedTime = (endNumber - startNumber) * 3000;
    BlockExtention startNumberBlock = apiWrapper.getBlockByNum(startNumber);
    BlockExtention endNumberBlock = apiWrapper.getBlockByNum(endNumber);
    long actualTime = endNumberBlock.getBlockHeader().getRawData().getTimestamp() - startNumberBlock
        .getBlockHeader().getRawData().getTimestamp();
    logger.info("expectedTime: {}, actual time: {}", expectedTime, actualTime);

    float tps = (float) (1.0 * totalTrxCnt * 1000 / actualTime);
    float missBlockRate = (float) (1.0 * (actualTime - expectedTime) / actualTime);

    logger.info("Stress test report:");
    logger.info("broadcast tps limit: {}", broadcastTpsLimit);
    logger.info("statistic block range: startBlock: {}, endBlock: {}", startNumber, endNumber);
    logger.info("total generate tx count: {}, total broadcast tx count: {}, tx on chain rate: {}",
        totalGenerateTxCnt, totalTrxCnt, 1.0 * totalTrxCnt / totalGenerateTxCnt);
    logger.info("cost time: {} minutes", 1.0 * actualTime / (60 * 1000));
    logger.info("max block size: {}", maxTrxCntInOneBlock);
    logger.info("min block size: {}", minTrxCntInOneBlock);
    logger.info("tps: {}", tps);
    logger.info("miss block rate: {}", missBlockRate);

    try (BufferedWriter writer = new BufferedWriter(new FileWriter(output))) {
      writer.write("Stress test report:");
      writer.newLine();
      writer.write(String.format("broadcast tps limit: %d", broadcastTpsLimit));
      writer.newLine();
      writer.write(String
          .format("statistic block range: startBlock: %d, endBlock: %d", startNumber, endNumber));
      writer.newLine();
      writer.write(String
          .format("total generate tx count: %d, total broadcast tx count: %d, tx on chain rate: %f",
              totalGenerateTxCnt, totalTrxCnt, 1.0 * totalTrxCnt / totalGenerateTxCnt));
      writer.newLine();
      writer.write(String.format("cost time: %f minutes", 1.0 * actualTime / (60 * 1000)));
      writer.newLine();
      writer.write(String.format("max block size: %d", maxTrxCntInOneBlock));
      writer.newLine();
      writer.write(String.format("min block size: %d", minTrxCntInOneBlock));
      writer.newLine();
      writer.write(String.format("tps: %f", tps));
      writer.newLine();
      writer.write(String.format("miss block rate: %f", missBlockRate));
    } catch (IOException e) {
      e.printStackTrace();
    }
  }
}
