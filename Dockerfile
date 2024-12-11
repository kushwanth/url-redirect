FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod tidy
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o url-redirect .

FROM gcr.io/distroless/static:latest

COPY --from=builder /app/url-redirect .
EXPOSE 8082

CMD ["./url-redirect"]
