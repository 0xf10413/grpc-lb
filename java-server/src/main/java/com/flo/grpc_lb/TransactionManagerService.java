package com.flo.grpc_lb;

import java.util.logging.Logger;

public class TransactionManagerService extends TransactionManagerGrpc.TransactionManagerImplBase {
    private static Logger logger = Logger.getLogger(TransactionManagerService.class.getCanonicalName());
    ClientCounterFilterService clientCounter;
    private final int maxTransactions;

    public TransactionManagerService(ClientCounterFilterService clientCounter, int maxTransactions) {
        this.clientCounter = clientCounter;
        this.maxTransactions = maxTransactions;
    }
    
    public void startTransaction(com.flo.grpc_lb.Query query,
        io.grpc.stub.StreamObserver<com.flo.grpc_lb.Reply> responseObserver) {
            logger.info("Got a query with id " + query.getId() + "!");
            logger.info("Btw there are " + clientCounter.getNbClients() + " clients connected");

            Reply reply = Reply.newBuilder()
                .setDisconnect(query.getId() >= maxTransactions ? -1 : 10)
                .build();
            responseObserver.onNext(reply);
            responseObserver.onCompleted();
        }
}
