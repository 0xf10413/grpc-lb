package com.flo.grpclb;

import java.util.logging.Logger;

import io.grpc.Context;
import io.grpc.stub.StreamObserver;

public class TransactionManagerService extends TransactionManagerGrpc.TransactionManagerImplBase {
    private static Logger logger = Logger.getLogger(TransactionManagerService.class.getCanonicalName());
    ClientCounterFilterService clientCounter;
    private final int maxConnectionDurationSeconds;

    public TransactionManagerService(ClientCounterFilterService clientCounter, int maxTransactions) {
        this.clientCounter = clientCounter;
        this.maxConnectionDurationSeconds = maxTransactions;
    }

    private void computeClientMustDisconnect(ClientLease lease) {
        synchronized(clientCounter) {
            long maxClients = clientCounter.getMaxNbClients();
            long nbClients = clientCounter.getNbActuallyConnectedClients();

            if (maxClients >= 0 && nbClients > maxClients) {
                logger.info((nbClients - maxClients) + " too many clients ! One will be asked to leave.");
                lease.expire();
            }
        }
    }
    
    public void startTransaction(Query query,
        StreamObserver<Reply> responseObserver) {
            logger.info("Got a query with id " + query.getId() + "!");
            logger.info("Btw there are " + clientCounter.getNbActuallyConnectedClients() + " clients connected");

            ClientLease lease = TransactionManagerInterceptor.clientLeaseKey.get();
            long clientConnectedDurationSeconds = (System.currentTimeMillis() - lease.getActualStartTimestamp())/1000;
            logger.info("Btw client has been connected for  " + clientConnectedDurationSeconds + "s");
            
            computeClientMustDisconnect(lease);
            lease.renew(); // TODO: hacky…

            Reply reply = Reply.newBuilder()
                .setDisconnect(lease.expired() ? -1 : 20)
                .build();
            responseObserver.onNext(reply);
            responseObserver.onCompleted();
        }

}
