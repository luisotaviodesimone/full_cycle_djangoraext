from django import forms
from django.core.files import File

MAX_VIDEO_CHUNK_SIZE = 1 * 1024 * 1024  # 1MB


class VideoChunkUploadForm(forms.Form):
    chunk = forms.FileField(required=True)
    chunkIndex = forms.IntegerField(min_value=0, required=True)

    def clean_chunk(self):
        chunk = self.cleaned_data.get("chunk")

        # if type(chunk) == File:
        if chunk.size > MAX_VIDEO_CHUNK_SIZE:
            raise forms.ValidationError("O arquivo deve ser um vídeo no formato MP4.")

        return chunk


class VideoChunkFinishUploadForm(forms.Form):
    fileName = forms.CharField(max_length=255, required=True)
    totalChunks = forms.IntegerField(min_value=1, required=True)
