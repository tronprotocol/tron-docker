package org.tron;

import com.typesafe.config.Config;
import com.typesafe.config.ConfigFactory;
import java.io.File;
import java.io.IOException;
import java.nio.file.Paths;
import java.util.concurrent.Callable;
import lombok.extern.slf4j.Slf4j;
import org.tron.common.application.Application;
import org.tron.common.application.ApplicationFactory;
import org.tron.common.application.TronApplicationContext;
import org.tron.core.config.DefaultConfig;
import org.tron.core.config.args.Args;
import org.tron.core.net.TronNetDelegate;
import org.tron.core.net.TronNetService;
import org.tron.core.services.RpcApiService;
import org.tron.trident.core.ApiWrapper;
import org.tron.trident.core.exceptions.IllegalException;
import org.tron.trident.proto.Chain.Block;
import org.tron.trxs.BroadcastGenerate;
import org.tron.trxs.BroadcastRelay;
import org.tron.trxs.TxConfig;
import org.tron.utils.PublicMethod;
import org.tron.utils.Statistic;
import picocli.CommandLine;
import picocli.CommandLine.Command;

@Slf4j(topic = "broadcast")
@Command(name = "broadcast",
    description = "Broadcast the transactions and compute the TPS.",
    exitCodeListHeading = "Exit Codes:%n",
    exitCodeList = {
        "0:Successful",
        "n:Internal error: exception occurred, please check logs/stress_test.log"})
public class BroadcastTx implements Callable<Integer> {

  @CommandLine.Spec
  public static CommandLine.Model.CommandSpec spec;

  @CommandLine.Option(names = {"-c", "--config"},
      defaultValue = "stress.conf",
      description = "configure the parameters for broadcasting transactions."
          + " Default: ${DEFAULT-VALUE}")
  private String config;

  @CommandLine.Option(names = {"-d", "--database-directory"},
      defaultValue = "output-directory",
      description = "java-tron database directory path. Default: ${DEFAULT-VALUE}")
  private String database;

  @CommandLine.Option(names = {"--fn-config"},
      defaultValue = "config.conf",
      description = "configure the parameters for broadcasting transactions."
          + " Default: ${DEFAULT-VALUE}")
  private String fnConfig;

  @CommandLine.Option(names = {"-h", "--help"})
  private boolean help;

  private ApiWrapper apiWrapper;

  private TronApplicationContext context;
  private Application app;
  private TronNetService tronNetService;
  private TronNetDelegate tronNetDelegate;

  @Override
  public Integer call() throws IOException, InterruptedException, IllegalException {
    if (help) {
      spec.commandLine().usage(System.out);
      return 0;
    }

    Config stressConfig = ConfigFactory.load();
    File file = Paths.get(config).toFile();
    if (file.exists() && file.isFile()) {
      stressConfig = ConfigFactory.parseFile(Paths.get(config).toFile());
    } else {
      logger.error("Stress test config file [" + config + "] not exists!");
      spec.commandLine().getErr()
          .format("Stress test config file [%s] not exists!", config)
          .println();
      System.exit(1);
    }
    TxConfig.initParams(stressConfig);
    TxConfig config = TxConfig.getInstance();

    File fnConfigFile = Paths.get(fnConfig).toFile();
    if (!fnConfigFile.exists() || fnConfigFile.isDirectory()) {
      logger.error("full node config file [" + fnConfig + "] not exists!");
      spec.commandLine().getErr()
          .format("full node config file [%s] not exists!", fnConfig)
          .println();
      System.exit(1);
    }

    Args.setParam(new String[]{"-d", database}, fnConfig);
    int rpcPort = PublicMethod.chooseRandomPort();
    Args.getInstance().setRpcPort(rpcPort);
    int nodeListenPort = PublicMethod.chooseRandomPort();
    while (nodeListenPort == rpcPort) {
      nodeListenPort = PublicMethod.chooseRandomPort();
    }
    Args.getInstance().setNodeListenPort(nodeListenPort);
    logger.info("rpc.port: {}, node.listen.port {}", rpcPort, nodeListenPort);

    context = new TronApplicationContext(DefaultConfig.class);
    app = ApplicationFactory.create(context);
    app.addService(context.getBean(RpcApiService.class));
    app.startup();

    String url = String.format("%s:%d", "127.0.0.1",
        Args.getInstance().getRpcPort());
    apiWrapper = new ApiWrapper(url, url, config.getPrivateKey());
    Statistic.setApiWrapper(apiWrapper);
    Statistic.setBroadcastTpsLimit(config.getBroadcastTpsLimit());
    Statistic.setTotalGenerateTxCnt(config.getTotalTxCnt());

    tronNetDelegate = context.getBean(TronNetDelegate.class);
    while (getBroadCastPeerCount() <= 0) {
      logger.warn("No available peers to broadcast, please wait");
      Thread.sleep(1000);
    }

    tronNetService = context.getBean(TronNetService.class);
    if (config.isBroadcastGenerate()) {
      BroadcastGenerate broadcastGenerate = new BroadcastGenerate(config, tronNetService);
      Block startBlock = apiWrapper.getNowBlock();
      broadcastGenerate.broadcastTransactions();
      Block endBlock = apiWrapper.getNowBlock();
      long startNumber = startBlock.getBlockHeader().getRawData().getNumber();
      long endNumber = endBlock.getBlockHeader().getRawData().getNumber();
      Statistic.result(startNumber, endNumber, "stress-test-output/broadcast-generate-result");
    }

    if (config.isBroadcastRelay()) {
      BroadcastRelay broadcastRelay = new BroadcastRelay(tronNetService);
      Block startBlock = apiWrapper.getNowBlock();
      broadcastRelay.broadcastTransactions();
      Block endBlock = apiWrapper.getNowBlock();
      long startNumber = startBlock.getBlockHeader().getRawData().getNumber();
      long endNumber = endBlock.getBlockHeader().getRawData().getNumber();
      Statistic.result(startNumber, endNumber, "stress-test-output/broadcast-relay-result");
    }

    shutdown();
    return 0;
  }

  private int getBroadCastPeerCount() {
    return (int) tronNetDelegate.getActivePeer().stream()
        .filter(p -> !p.isNeedSyncFromPeer()  && !p.isNeedSyncFromUs())
        .count();
  }

  private void shutdown() {
    context.close();
  }
}
