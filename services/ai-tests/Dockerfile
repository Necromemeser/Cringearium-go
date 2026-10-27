FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /ai-tests \
    ./services/ai-tests/cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /ai-tests /app/ai-tests

EXPOSE 8085

CMD ["/app/ai-tests"]