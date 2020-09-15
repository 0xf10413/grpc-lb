package com.flo.grpc_lb;

import io.grpc.Attributes;
import io.grpc.ServerTransportFilter;

import java.util.concurrent.atomic.AtomicInteger;
import java.util.logging.Logger;

import io.prometheus.client.Counter;

public class ClientCounterFilterService extends ServerTransportFilter {
    private static Logger logger = Logger.getLogger(ClientCounterFilterService.class.getCanonicalName());
    private AtomicInteger nbClients = new AtomicInteger();
    private static final Counter clientsConnectedCounter = Counter.build()
                        .name("clients_connected").help("Number of clients connected").register();
    private static final Counter clientsDisonnectedCounter = Counter.build()
                        .name("clients_disconnected").help("Number of clients disconnected").register();

    public Attributes transportReady(Attributes transportAttrs) {
        nbClients.incrementAndGet();
        clientsConnectedCounter.inc();
        logger.info("Transport is ready for someone ! There are now " + nbClients.get() + " clients.");
        return transportAttrs;
    }

    public void transportTerminated(Attributes transportAttrs) {
        nbClients.decrementAndGet();
        clientsDisonnectedCounter.inc();
        logger.info("Transport is terminated for someone ! There are now " + nbClients.get() + " clients.");
    }

    public int getNbClients() {
        return nbClients.get();
    }
}
