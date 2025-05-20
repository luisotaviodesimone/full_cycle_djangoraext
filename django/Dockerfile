FROM ghcr.io/astral-sh/uv:alpine

WORKDIR /app

COPY pyproject.toml uv.lock .

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONFAULTHANDLER 1
ENV PIPENV_VENV_IN_PROJECT 1
ENV UV_PROJECT_ENVIRONMENT /root/.venv

RUN uv sync

CMD source /root/.venv/bin/activate && python manage.py runserver 0.0.0.0:8000
