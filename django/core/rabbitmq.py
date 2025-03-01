from kombu import Connection
from videos import settings


def create_rabbitmq_connection() -> Connection:

    rabbitmq_url = settings.ENVIRONMENT["RABBITMQ_URL"]

    connection = Connection(str(rabbitmq_url))

    print("connection made")

    return connection
