from typing import Any

from django.contrib import admin, messages
from django.contrib.auth.admin import csrf_protect_m
from django.http import HttpRequest, JsonResponse
from django.shortcuts import render
from django.urls import path, reverse
from django.utils.html import format_html

from core.form import VideoChunkFinishUploadForm, VideoChunkUploadForm
from core.models import Tag, Video
from core.services import (
    VideoChunkUploadException,
    VideoMediaInvalidStatusException,
    VideoMediaNotExistsException,
    create_video_service_factory,
)


class VideoAdmin(admin.ModelAdmin):
    list_display = (
        "title",
        "published_at",
        "is_published",
        "num_likes",
        "num_views",
        "redirect_to_upload",
    )

    def video_status(self, obj: Video) -> str:
        return obj.get_video_status_display()

    def get_readonly_fields(self, request: HttpRequest, obj: Any | None) -> list[str]:
        return (
            # if video is being created
            [
                "video_status",
                "is_published",
                "published_at",
                "num_likes",
                "num_views",
                "author",
            ]
            if not obj
            # if video is being edited
            else ["video_status", "published_at", "num_likes", "num_views", "author"]
        )

    def get_urls(self):
        urls = super().get_urls()
        custom_urls = [
            path(
                "<int:id>/upload-video",
                self.admin_site.admin_view(self.upload_video_view),
                name="core_video_upload",
            ),
            path(
                "<int:id>/upload-video/finish",
                self.admin_site.admin_view(self.finish_upload_video),
                name="core_video_upload_finish",
            ),
        ]

        return custom_urls + urls

    def save_model(self, request: HttpRequest, obj, form, change) -> None:
        if not obj.pk:
            obj.author = request.user
        super().save_model(request, obj, form, change)

    @csrf_protect_m
    def upload_video_view(self, request, id):

        str_id = str(id)

        if request.method == "POST":
            return self._do_upload_video_chunks(request, id)

        try:
            video = create_video_service_factory().find_video(id)
            context = dict(
                self.admin_site.each_context(request),
                opts=self.model._meta,
                id=id,
                video=video,
                video_media=(
                    video.video_media if hasattr(video, "video_media") else None
                ),
                has_view_permission=True,
            )
            return render(request, "admin/core/upload_video.html", context)
        except:
            return self._get_obj_does_not_exist_redirect(request, self.opts, str_id)

    def _do_upload_video_chunks(self, request: HttpRequest, id: int) -> Any:
        form = VideoChunkUploadForm(request.POST, request.FILES)

        if not form.is_valid():
            return JsonResponse({"error": form.errors}, status=400)

        try:
            create_video_service_factory().process_upload(
                video_id=id,
                chunk_index=form.cleaned_data["chunkIndex"],
                chunk=form.cleaned_data["chunk"].read(),
            )
        except Video.DoesNotExist:
            return JsonResponse({"error": "Video not found"}, status=404)
        except Exception as e:
            return JsonResponse({"error": str(e)}, status=500)

        return JsonResponse({"message": "Chunk uploaded"}, status=200)

    def redirect_to_upload(self, obj: Video):
        url = reverse("admin:core_video_upload", args=[obj.id])
        return format_html(f'<a href="{url}">Upload</a>')

    def finish_upload_video(self, request, id):
        if request.method != "POST":
            return JsonResponse({"error": "Invalid method"}, status=405)

        form = VideoChunkFinishUploadForm(request.POST)

        if not form.is_valid():
            return JsonResponse({"error": form.errors}, status=400)

        try:
            video_service = create_video_service_factory()
            video_service.finalize_upload(
                video_id=id,
                total_chunks=form.cleaned_data["totalChunks"],
            )
        except Video.DoesNotExist:
            return JsonResponse({"error": "Video not found"}, status=404)
        except (
            VideoMediaNotExistsException,
            VideoMediaInvalidStatusException,
            VideoChunkUploadException,
        ) as e:
            return JsonResponse({"error": str(e)}, status=400)

        self.message_user(request, "Upload realizado com sucesso.", messages.SUCCESS)
        return JsonResponse({"message": "Upload finished"}, status=200)


admin.site.register(Video, VideoAdmin)
admin.site.register(Tag)
