FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /gateway \
    ./services/gateway/cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /gateway /app/gateway

EXPOSE 8080

CMD ["/app/gateway"]
