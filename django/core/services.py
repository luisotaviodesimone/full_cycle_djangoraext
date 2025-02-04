import os
import shutil
from dataclasses import dataclass

from django.db import IntegrityError, transaction

from core.models import Video, VideoMedia


class VideoMediaInvalidStatusException(Exception):
    pass


class VideoMediaNotExistsException(Exception):
    pass


class VideoChunkUploadException(Exception):
    pass


@dataclass
class VideoService:
    storage: "Storage"

    def get_chunk_directory(self, video_id: int) -> str:
        return f"../videos/{video_id}"

    def find_video(self, video_id: int) -> Video:
        return Video.objects.get(id=video_id)

    def __prepare_video_media(self, video: Video):
        try:
            video_media = video.video_media
        except Video.video_media.RelatedObjectDoesNotExist:
            try:
                video_media = VideoMedia.objects.create(
                    video=video,
                    status=VideoMedia.Status.UPLOADED_STARTED,
                    video_path=self.get_chunk_directory(video.id),
                )
            except IntegrityError:
                video_media = VideoMedia.objects.get(video=video)
        return video_media

    def process_upload(self, video_id: int, chunk_index: int, chunk: bytes) -> None:
        video = self.find_video(video_id)
        directory = self.get_chunk_directory(video_id)

        # Atualiza o status para 'UPLOADED_STARTED' e armazena o chunk
        sid = transaction.savepoint()
        try:
            video_media = self.__prepare_video_media(video)

            if (
                video_media.status == VideoMedia.Status.PROCESS_STARTED
            ):  # upload finalizado
                raise VideoMediaInvalidStatusException(
                    "An upload is already in progress."
                )

            # Se o status já está em 'UPLOADED_STARTED', continua armazenando os chunks
            if video_media.status == VideoMedia.Status.UPLOADED_STARTED:
                self.storage.storage_chunk(
                    str(video_media.video_path), chunk_index, chunk
                )
                return

            # Se o processamento já terminou, reinicia o caminho do diretório de chunks
            if video_media.status == VideoMedia.Status.PROCESS_FINISHED:
                video_media.video_path = directory
                video.is_published = False
                video.save()

            video_media.status = VideoMedia.Status.UPLOADED_STARTED
            video_media.save()
            self.storage.storage_chunk(str(video_media.video_path), chunk_index, chunk)
            transaction.savepoint_commit(sid)
        except Exception as e:
            transaction.savepoint_rollback(sid)
            raise e

    def finalize_upload(self, video_id: int, total_chunks: int):
        video = self.find_video(video_id)

        try:
            video_media = video.video_media
            if video_media.status != VideoMedia.Status.UPLOADED_STARTED:
                raise VideoMediaInvalidStatusException(
                    "Upload must be started to finish it."
                )
            is_chunks_valid = self.__validate_chunks(
                str(video.video_media.video_path), total_chunks
            )
            if not is_chunks_valid:
                raise VideoChunkUploadException("Chunks are invalid.")

            video_media.status = VideoMedia.Status.PROCESS_STARTED
            video_media.save()

            # self.__produce_message(video_id, video_media.video_path, "chunks")

        except Video.video_media.RelatedObjectDoesNotExist:
            raise VideoMediaNotExistsException("Upload not started.")

    def __validate_chunks(self, video_path: str, total_chunks: int) -> bool:
        if not os.path.exists(video_path):
            return False

        for i in range(total_chunks):
            chunk_path = os.path.join(video_path, f"{i}.chunk")
            if not os.path.exists(chunk_path):
                return False

        return True

def create_video_service_factory() -> VideoService:
    return VideoService(Storage())

class Storage:

    def storage_chunk(self, directory: str, chunk_index: int, chunk: bytes) -> None:
        if not os.path.exists(directory):
            os.makedirs(directory)

        chunk_path = os.path.join(directory, f"{chunk_index}.chunk")

        with open(chunk_path, "wb") as chunk_file:
            chunk_file.write(chunk)

    def move_chunks(self, source_path: str, dest_path: str) -> None:

        if not os.path.exists(dest_path):
            os.makedirs(dest_path, exist_ok=True)

        for filename in os.listdir(source_path):
            file_path = os.path.join(source_path, filename)

            if os.path.isfile(file_path):
                try:
                    shutil.move(file_path, os.path.join(dest_path, filename))
                    print(f"Arquivo {source_path}/{filename} movido para {dest_path}.")
                except Exception as e:
                    print(f"Erro ao mover o arquivo {source_path}/{filename}: {e}")
            else:
                print(f"{file_path} não é um arquivo.")
