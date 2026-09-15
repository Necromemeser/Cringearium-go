FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /auth \
    ./services/auth/cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /auth /app/auth

EXPOSE 8083

CMD ["/app/payment"]