# Stage 1: Build the Go binary
FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/blog-api main.go

FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/blog-api /app/blog-api

RUN chmod +x /app/blog-api

EXPOSE 8080

CMD ["./blog-api"]
