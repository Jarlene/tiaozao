FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download && go build -o server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ["./server"]
