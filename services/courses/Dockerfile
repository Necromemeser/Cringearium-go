FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /courses \
    ./services/courses/cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /courses /app/courses

EXPOSE 8082

CMD ["/app/courses"]