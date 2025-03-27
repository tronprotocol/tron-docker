package org.tron;


import com.typesafe.config.Config;
import com.typesafe.config.ConfigFactory;
import java.io.File;
import java.nio.file.Paths;
import java.util.concurrent.Callable;
import lombok.extern.slf4j.Slf4j;
import org.tron.trident.core.ApiWrapper;
import org.tron.trxs.TxConfig;
import org.tron.utils.Statistic;
import picocli.CommandLine;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;

@Slf4j(topic = "statistic")
@Command(name = "statistic",
    description = "Compute the TPS from specified block range.",
    exitCodeListHeading = "Exit Codes:%n",
    exitCodeList = {
        "0:Successful",
        "n:Internal error: exception occurred, please check logs/stress_test.log"})
public class TpsStatistic implements Callable<Integer> {

  @CommandLine.Spec
  public static CommandLine.Model.CommandSpec spec;

  @CommandLine.Option(names = {"-c", "--config"},
      defaultValue = "stress.conf",
      description = "configure the parameters for broadcasting transactions."
          + " Default: ${DEFAULT-VALUE}")
  private String config;

  @CommandLine.Option(names = {"-h", "--help"})
  private boolean help;

  private long startBlock;
  private long endBlock;

  @CommandLine.Option(names = {"-o", "--output"},
      defaultValue = "tps-statistic-result",
      description = "store the tps statistic result."
          + " Default: ${DEFAULT-VALUE}")
  private String output;

  @Override
  public Integer call() throws Exception {
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

    startBlock = config.getStatisticStartNumber();
    endBlock = config.getStatisticEndNumber();
    if (startBlock < 0 || startBlock >= endBlock) {
      logger.error("invalid start Block: {}, end Block: {}", startBlock, endBlock);
      spec.commandLine().getErr()
          .format("invalid start Block: %d, end Block: %d", startBlock, endBlock)
          .println();
      System.exit(1);
    }

    if (config.getStatisticUrl().isEmpty()) {
      logger.error("no available broadcast url found.");
      spec.commandLine().getErr().println("no available broadcast url found.");
      System.exit(1);
    }

    Statistic.setApiWrapper(
        new ApiWrapper(config.getStatisticUrl(), config.getStatisticUrl(),
            config.getPrivateKey()));
    Statistic.result(startBlock, endBlock, output);
    return 0;
  }
}
