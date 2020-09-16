package com.flo.grpclb;

import java.io.IOException;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.prometheus.client.exporter.HTTPServer;

public class App {
    public static void main(String[] args) throws InterruptedException, IOException {
        final String maxTransactionsStr = System.getenv("MAX_TRANSACTIONS");
        final int maxTransactions;
        if (maxTransactionsStr == null) {
            maxTransactions = 0;
        } else {
            maxTransactions = Integer.valueOf(maxTransactionsStr);
        }

        // Prometheus setup
        HTTPServer prometheusServer = new HTTPServer(1234);
        prometheusServer.getPort();

        ClientCounterFilterService clientCounter = new ClientCounterFilterService();
        Server server = ServerBuilder.forPort(50051)
                    .addTransportFilter(clientCounter)
                    .addService(new TransactionManagerService(clientCounter, maxTransactions))
                    .addService(new LoadBalancerService(clientCounter))
                    .build();
        System.out.println("Starting server…");
        server.start();
        server.awaitTermination();
        System.out.println("Stopped server…");
    }
}
