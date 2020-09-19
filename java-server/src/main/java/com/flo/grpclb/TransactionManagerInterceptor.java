package com.flo.grpclb;

import io.grpc.Context;
import io.grpc.Contexts;
import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCall.Listener;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;

public class TransactionManagerInterceptor implements ServerInterceptor {
    public static final Context.Key<ClientLease> clientLeaseKey = Context.key("ClientLease");

    @Override
    public <ReqT, RespT> Listener<ReqT> interceptCall(ServerCall<ReqT, RespT> call, Metadata headers,
            ServerCallHandler<ReqT, RespT> next) {

        // Forward content of transport attributes to context values
        ClientLease lease = call.getAttributes().get(ClientCounterFilterService.clientLeaseKey);
        Context context = Context.current()
            .withValue(clientLeaseKey, lease);

        return Contexts.interceptCall(context, call, headers, next);
    }
    
}
