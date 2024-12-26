from django.contrib import admin
from core.models import Video, Tag

class VideoAdmin(admin.ModelAdmin):
    def get_urls(self):
        urls = super().get_urls()
        custom_urls = []
        return urls + custom_urls

admin.site.register(Video)
admin.site.register(Tag)
