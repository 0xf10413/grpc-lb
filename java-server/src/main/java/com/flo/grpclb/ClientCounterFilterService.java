package com.flo.grpclb;

import io.grpc.Attributes;
import io.grpc.Context;
import io.grpc.ServerTransportFilter;

import java.util.concurrent.atomic.AtomicInteger;
import java.util.logging.Logger;

import io.prometheus.client.Counter;

public class ClientCounterFilterService extends ServerTransportFilter {
    private static Logger logger = Logger.getLogger(ClientCounterFilterService.class.getCanonicalName());
    private AtomicInteger nbActuallyConnectedClients = new AtomicInteger();
    private AtomicInteger nbClients = new AtomicInteger(); // Clients that were not told to go away
    private int maxNbClients = -1;
    private final long leaseDuration;
    public static final Attributes.Key<ClientLease> clientLeaseKey = Attributes.Key.create("ClientLease");

    private static final Counter clientsConnectedCounter = Counter.build()
                        .name("clients_connected").help("Number of clients connected").register();
    private static final Counter clientsDisconnectedCounter = Counter.build()
                        .name("clients_disconnected").help("Number of clients disconnected").register();



    public ClientCounterFilterService(long leaseDuration) {
        this.leaseDuration = leaseDuration;
    }

    public Attributes transportReady(Attributes transportAttrs) {
        synchronized(this) {
            nbActuallyConnectedClients.incrementAndGet();
            nbClients.incrementAndGet();
            clientsConnectedCounter.inc();

            transportAttrs = transportAttrs.toBuilder()
                .set(clientLeaseKey, new ClientLease(leaseDuration))
                .build();

            logger.info("Transport is ready for someone ! There are now " + nbActuallyConnectedClients.get() + " clients.");
            return super.transportReady(transportAttrs);
        }
    }

    public void transportTerminated(Attributes transportAttrs) {
        synchronized(this) {
            ClientLease lease = transportAttrs.get(clientLeaseKey);

            // If the client disconnected by themselves we still need to count it
            if (!lease.forceExpired()) {
                nbClients.decrementAndGet();
            }
    
            nbActuallyConnectedClients.decrementAndGet();
            clientsDisconnectedCounter.inc();
            logger.info("Transport is terminated for someone ! " +
                "There are now " + nbActuallyConnectedClients.get() + " clients connected, " +
                "and " + nbClients.get() + " clients still considered.");
        }
    }

    public int getNbActuallyConnectedClients() {
        return nbActuallyConnectedClients.get();
    }

    public int getNbClients() {
        return nbClients.get();
    }

    public int getMaxNbClients() {
        return maxNbClients;
    }

    public void setMaxNbClients(int maxNbClients) {
        this.maxNbClients = maxNbClients;
    }

    public synchronized void flagClientToBeDisconnected() {
        this.nbClients.decrementAndGet();
    }
}
