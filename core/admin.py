from django.contrib import admin
from django.contrib.auth.admin import csrf_protect_m
from django.http import HttpResponse, JsonResponse
from django.http.response import responses
from django.shortcuts import render
from django.urls import path, reverse
from django.utils.html import format_html

from core.form import VideoChunkFinishUploadForm, VideoChunkUploadForm
from core.models import Tag, Video
from core.services import VideoService, create_video_service_factory


class VideoAdmin(admin.ModelAdmin):
    list_display = (
        "title",
        "published_at",
        "is_published",
        "num_likes",
        "num_views",
        "redirect_to_upload",
    )

    def get_urls(self):
        urls = super().get_urls()
        custom_urls = [
            path("<int:id>/upload-video", self.upload_video, name="core_video_upload"),
            path(
                "<int:id>/upload-video/finish",
                self.finish_upload_video,
                name="core_video_upload_finish",
            ),
        ]

        return custom_urls + urls

    @csrf_protect_m
    def upload_video(self, request, id):
        if request.method == "POST":
            form = VideoChunkUploadForm(request.POST, request.FILES)
            if not form.is_valid():
                return JsonResponse({"error": form.errors}, status=400)
            video_service = create_video_service_factory()
            video_service.process_upload(
                video_id=id,
                chunk_index=form.cleaned_data["chunkIndex"],
                chunk=form.cleaned_data["chunk"].read(),
            )

        context = dict(
            id=id,
        )

        return render(request, "admin/core/upload_video.html", context)

    def redirect_to_upload(self, obj: Video):
        url = reverse("admin:core_video_upload", args=[obj.id])
        return format_html(f'<a href="{url}">Upload</a>')

    def finish_upload_video(self, request, id):
        if request.method != "POST":
            return JsonResponse({"error": "Invalid method"}, status=405)

        form = VideoChunkFinishUploadForm(request.POST)

        if not form.is_valid():
            return JsonResponse({"error": form.errors}, status=400)

        video_service = create_video_service_factory()
        video_service.finalize_upload(
            video_id=id,
            total_chunks=form.cleaned_data["totalChunks"],
        )

        return JsonResponse({"message": "Upload finished"})


admin.site.register(Video, VideoAdmin)
admin.site.register(Tag)
