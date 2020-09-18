package com.flo.grpclb;

import java.util.logging.Logger;

import com.flo.grpclb.LoadBalancingManagerGrpc.LoadBalancingManagerImplBase;

import io.grpc.stub.StreamObserver;

public class LoadBalancerService extends LoadBalancingManagerImplBase {
    private static Logger logger = Logger.getLogger(LoadBalancerService.class.getCanonicalName());
    private ClientCounterFilterService clientCounter;

    public LoadBalancerService(ClientCounterFilterService clientCounter) {
        logger.info("Loadbalancer query received!");
        this.clientCounter = clientCounter;
    }

    @Override
    public void getClientStatus(ClientRequest request, StreamObserver<ClientStatus> responseObserver) {
        ClientStatus clientStatus = ClientStatus.newBuilder()
                .setNbClients(clientCounter.getNbActuallyConnectedClients())
                .setMaxNbClients(clientCounter.getMaxNbClients())
                .build();
        responseObserver.onNext(clientStatus);
        responseObserver.onCompleted();
    }

    @Override
    public void setMaxClients(SetMaxClientsRequest request, StreamObserver<SetMaxClientsReply> responseObserver) {
        SetMaxClientsReply maxClientsReply = SetMaxClientsReply.newBuilder().build();
        clientCounter.setMaxNbClients(request.getMaxNbClients());
        responseObserver.onNext(maxClientsReply);
        responseObserver.onCompleted();
    }
}
