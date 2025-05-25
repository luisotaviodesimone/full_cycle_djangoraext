from kombu import Connection, Exchange, Queue

from videos import settings


def create_rabbitmq_connection() -> Connection:

    rabbitmq_url = settings.ENVIRONMENT["RABBITMQ_URL"]

    connection = Connection(str(rabbitmq_url))

    return connection


def use_rabbitmq_queue(queue_name: str, exchange_name: str, routing_key: str) -> Queue:

    exchange = Exchange(exchange_name, auto_delete=True)
    queue = Queue(queue_name, exchange, routing_key=routing_key)

    return queue
