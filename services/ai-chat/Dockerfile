FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /ai-chat \
    ./services/ai-chat/cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /ai-chat /app/ai-chat

EXPOSE 8084

CMD ["/app/ai-chat"]