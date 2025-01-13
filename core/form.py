from django import forms

MAX_VIDEO_CHUNK_SIZE = 1 * 1024 * 1024 # 1MB

class VideoChunkUploadForm(forms.Form):
    chunk = forms.FileField(required=True)
