FROM python:3.12.9-alpine

WORKDIR /app

COPY Pipfile* .

ENV PIPENV_VENV_IN_PROJECT=1
 
RUN pip install pipenv

RUN pipenv install --system

CMD ["python3", "manage.py", "runserver", "0.0.0.0:8000"]
