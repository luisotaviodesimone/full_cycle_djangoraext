# Generated by Django 5.1.4 on 2024-12-26 15:53

from django.db import migrations


class Migration(migrations.Migration):

    dependencies = [
        ("core", "0001_initial"),
    ]

    operations = [
        migrations.RemoveField(
            model_name="video",
            name="author",
        ),
    ]
