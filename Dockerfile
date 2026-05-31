FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o thesis-app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates && adduser -D -s /bin/sh appuser
WORKDIR /app

COPY --from=builder /app/thesis-app .
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p uploads && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./thesis-app"]
