from core.rabbitmq import create_rabbitmq_connection, use_rabbitmq_queue
from core.services import create_video_service_factory
from videos import settings

from django.core.management import BaseCommand


class Command(BaseCommand):
    help = "Uploads chunks to external storage"

    def handle(self, *args, **options):
        self.stdout.write(self.style.SUCCESS("Starting consumer..."))
        exchange_name = str(settings.ENVIRONMENT["RABBITMQ_EXCHANGE"])
        routing_key = "chunks"
        queue = use_rabbitmq_queue(routing_key, exchange_name, routing_key)

        with create_rabbitmq_connection() as connection:
            with connection.Consumer(queue, callbacks=[self.process_message]):
                while True:
                    self.stdout.write(self.style.SUCCESS("Waiting for messages..."))
                    connection.drain_events()

    def process_message(self, body, message):
        self.stdout.write(self.style.SUCCESS(f"Processing message: {body}"))
        create_video_service_factory().upload_chunks_to_external_storage(body["video_id"])
        message.ack()
