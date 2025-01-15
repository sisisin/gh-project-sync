# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=bind,source=go.mod,target=go.mod \
  go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,target=. \
  go build -a -installsuffix cgo -o /bin/ghpsync -tags timetzdata ./cmd/sync

FROM debian:bookworm-slim AS final

WORKDIR /app

ARG UID=10001

RUN apt-get update && apt-get install -y ca-certificates openssl curl \
  && curl -sfS https://dotenvx.sh/install.sh | sh

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid "${UID}" \
  appuser
USER appuser

COPY --from=builder /bin/ghpsync .
COPY ./cmd/sync/.env .

EXPOSE 8080
ENTRYPOINT [ "dotenvx", "run", "--env-file=/app/.env", "--"]
CMD [ "/app/ghpsync" ]
