FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/producer ./cmd/producer

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/app /app/app
COPY --from=builder /app/bin/producer /app/producer

CMD ["/app/app"]