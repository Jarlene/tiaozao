FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY vendor/ vendor/
COPY go.mod go.sum ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY web/ web/
EXPOSE 8080
CMD ["./server"]
