package com.flo.grpc_lb;

import java.util.logging.Logger;

import com.flo.grpc_lb.LoadBalancingManagerGrpc.LoadBalancingManagerImplBase;

public class LoadBalancerService extends LoadBalancingManagerImplBase {
    private static Logger logger = Logger.getLogger(LoadBalancerService.class.getCanonicalName());
    private ClientCounterFilterService clientCounter;

    public LoadBalancerService(ClientCounterFilterService clientCounter) {
        this.clientCounter = clientCounter;
    }
}
