FROM golang:1.27-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN set -eu; mkdir /out; \
    for service in auth courses ai-chat ai-tests gateway; do \
      CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
        -o "/out/$service" "./services/$service/cmd/api"; \
    done

FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM debian:bookworm-slim AS runtime
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      ca-certificates nginx supervisor postgresql-client python3 && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /out/ /app/bin/
COPY --from=frontend /src/frontend/dist/ /usr/share/nginx/html/
COPY deploy/bundled/ /app/deploy/
RUN mkdir -p /run/nginx /run/cringearium && \
    chmod -R a+rX /usr/share/nginx/html && \
    cp /app/deploy/nginx.conf /etc/nginx/nginx.conf

ENV POSTGRES_HOST=postgres \
    POSTGRES_PORT=5432 \
    POSTGRES_USER=cringearium \
    POSTGRES_SSLMODE=disable \
    POSTGRES_WAIT_TIMEOUT=120 \
    JWT_TTL=24h \
    AUTH_URL=http://127.0.0.1:8081 \
    COURSES_URL=http://127.0.0.1:8082 \
    AI_CHAT_URL=http://127.0.0.1:8084 \
    AI_TESTS_URL=http://127.0.0.1:8085 \
    SERVER_ADDR=127.0.0.1:8080

EXPOSE 80
STOPSIGNAL SIGTERM
HEALTHCHECK --interval=30s --timeout=10s --start-period=180s --retries=3 \
    CMD python3 /app/deploy/healthcheck.py --check
ENTRYPOINT ["python3", "/app/deploy/entrypoint.py"]
