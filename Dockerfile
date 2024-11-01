FROM golang:1.23 AS builder

WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod tidy
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o url-redirect .

FROM alpine:latest

WORKDIR /root/
RUN wget https://github.com/P3TERX/GeoLite.mmdb/releases/download/2024.11.01/GeoLite2-Country.mmdb
COPY --from=builder /app/url-redirect .
EXPOSE 8082

CMD ["./url-redirect"]
