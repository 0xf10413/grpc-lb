import logging
import concurrent
import time
import os
import threading

import grpc
logging.basicConfig(format="[%(asctime)s] { %(threadName)s "
                          "%(filename)s:%(lineno)d} "
                          "%(levelname)s - %(message)s")
logging.getLogger().setLevel(logging.DEBUG)


import transaction_pb2
import transaction_pb2_grpc

MAX_TRANSACTIONS = int(os.getenv("MAX_TRANSACTIONS", 50))


class TransactionServer(transaction_pb2_grpc.TransactionManagerServicer):

    def StartTransaction(self, request, context):
        logging.info("Got a query with id %s from %s!", request.id, context.peer())
        reply = transaction_pb2.Reply()
        if request.id >= MAX_TRANSACTIONS:
            logging.info("Requesting disconnection")
            reply.disconnect = True
        return reply

if __name__ == "__main__":
    hostport = '[::]:50051'

    server = grpc.server(concurrent.futures.ThreadPoolExecutor(max_workers=10))

    transaction_pb2_grpc.add_TransactionManagerServicer_to_server(TransactionServer(), server)
    server.add_insecure_port(hostport)
    server.start()
    logging.info("Serving on %s", hostport)
    server.wait_for_termination()
