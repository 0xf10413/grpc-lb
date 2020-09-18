package com.flo.grpclb;

import java.io.IOException;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.ServerInterceptor;
import io.grpc.ServerInterceptors;
import io.prometheus.client.exporter.HTTPServer;

public class App {
    public static void main(String[] args) throws InterruptedException, IOException {
        final String maxConnectionDurationStr = System.getenv("MAX_CONNECTION_DURATION");
        final int maxConnectionDuration;
        if (maxConnectionDurationStr == null) {
            maxConnectionDuration = 0;
        } else {
            maxConnectionDuration = Integer.valueOf(maxConnectionDurationStr);
        }

        final long leaseDuration = 20;

        // Prometheus setup
        HTTPServer prometheusServer = new HTTPServer(1234);
        prometheusServer.getPort();

        ClientCounterFilterService clientCounter = new ClientCounterFilterService(leaseDuration);
        Server server = ServerBuilder.forPort(50051)
                    .addTransportFilter(clientCounter)
                    .addService(ServerInterceptors.intercept(
                        new TransactionManagerService(clientCounter, maxConnectionDuration),
                        new TransactionManagerInterceptor()))
                    .build();
        Server lbServer = ServerBuilder.forPort(50052)
                    .addService(new LoadBalancerService(clientCounter))
                    .build();
        System.out.println("Starting servers…");
        lbServer.start();
        server.start();
        server.awaitTermination();
        System.out.println("Stopped servers…");
    }
}
