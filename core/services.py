from dataclasses import dataclass

from core.models import Video


@dataclass
class VideoService:
    def find_video(self, video_id: int) -> Video:
        return Video.objects.get(id=video_id)

    def process_upload(self, video_id: int, chunk_index: int, chunk: bytes) -> None:
        print(f"video_id: {video_id}\nchunk: {chunk_index}")
        pass

    def finalize_upload(self, video_id: int, total_chunks: int):
        pass
