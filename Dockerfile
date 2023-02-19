FROM python:3.10-slim

RUN apt update -y \
    && apt install -y curl

ENV POETRY_HOME="/opt/poetry" \
    POETRY_VERSION=1.2.2

ENV PATH="$POETRY_HOME/bin:$PATH"

RUN curl -sSL https://install.python-poetry.org | python3 -

ADD . ./

RUN poetry config virtualenvs.create false && poetry install

EXPOSE 25565
ENTRYPOINT ["poetry", "run", "python3", "-m", "router"]
