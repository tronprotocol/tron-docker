package org.tron.utils;

import java.io.IOException;
import java.net.InetAddress;
import java.net.Socket;
import java.util.Random;

public class PublicMethod {

  public static int chooseRandomPort() {
    return chooseRandomPort(10240, 65000);
  }

  public static int chooseRandomPort(int min, int max) {
    int port = new Random().nextInt(max - min + 1) + min;
    try {
      while (!checkPortAvailable(port)) {
        port = new Random().nextInt(max - min + 1) + min;
      }
    } catch (IOException e) {
      return new Random().nextInt(max - min + 1) + min;
    }
    return port;
  }

  private static boolean checkPortAvailable(int port) throws IOException {
    InetAddress theAddress = InetAddress.getByName("127.0.0.1");
    try (Socket socket = new Socket(theAddress, port)) {
      // only check
      socket.getPort();
    } catch (IOException e) {
      return true;
    }
    return false;
  }

}
