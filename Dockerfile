FROM golang:1.21 AS builder

WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod tidy
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o url-redirect .

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/url-redirect .
EXPOSE 8082

CMD ["./url-redirect"]
