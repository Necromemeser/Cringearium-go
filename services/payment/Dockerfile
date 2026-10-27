FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /payment \
    ./services/payment/cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /payment /app/payment

EXPOSE 8083

CMD ["/app/payment"]