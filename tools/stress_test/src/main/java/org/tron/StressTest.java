package org.tron;

import java.util.concurrent.Callable;
import picocli.CommandLine;

@CommandLine.Command(subcommands = {CommandLine.HelpCommand.class, GenerateTx.class,
    BroadcastTx.class, TpsStatistic.class, CollectAddress.class})
public class StressTest implements Callable<Integer> {

  public static void main(String[] args) {
    CommandLine cli = new CommandLine(new StressTest());
    if (args == null || args.length == 0) {
      cli.usage(System.out);
    } else {
      int exitCode = cli.execute(args);
      System.exit(exitCode);
    }
  }

  @Override
  public Integer call() throws Exception {
    return 0;
  }
}
