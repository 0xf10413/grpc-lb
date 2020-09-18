import logging
import itertools
import time
import random

import grpc
logging.getLogger().setLevel(logging.DEBUG)
logging.basicConfig(format="[%(asctime)s] { %(threadName)s "
                          "%(filename)s:%(lineno)d} "
                          "%(levelname)s - %(message)s")

import transaction_pb2
import transaction_pb2_grpc

INSTANCE_ID = str(random.randint(1, 2))
METADATA = [('instance-id', INSTANCE_ID)]

class DisconnectException(Exception):
    """
    Exception class to disconnect from a server
    """
    pass

def randomCycleIterator(it):
    """
    Given a finite iterable :it,
    generates a random cycle iterator on it.

    Example: it=range(4), this could yield 3, 1, 0, 1, 2, 3, 1, …
    """
    values = list(it)
    while True:
        yield random.choice(values)

for hostport in itertools.cycle(['localhost:50051', '192.168.39.101:31044', 'localhost:50053']):
    time.sleep(1)
    reco_deadline = time.time() + 2
    logging.info("Will connect to hostport %s for %fs", hostport, reco_deadline - time.time())
    try:
        with grpc.insecure_channel(hostport) as channel:
            stub = transaction_pb2_grpc.TransactionManagerStub(channel)
            for i in itertools.count():
                logging.info("Sending query #%s", i)
                reply = stub.StartTransaction(transaction_pb2.Query(id=i), metadata=METADATA)
                if reply.disconnect != 0:
                    reco_deadline = time.time() + reply.disconnect
                    logging.info("Reconnection time updated to %ss in the future", reco_deadline - time.time())
                if reco_deadline - time.time() <= 0:
                    logging.info("Disconnection requested!")
                    raise DisconnectException()
                time.sleep(1)
    except DisconnectException:
        pass # Nothing special to do
    except Exception as e:
        logging.error("Got exception %s!", e)
