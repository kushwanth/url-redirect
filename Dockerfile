FROM golang:1.24-alpine AS builder

ARG TARGETARCH

WORKDIR /app
COPY app/go.mod app/go.sum ./
RUN go mod tidy
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o url-redirect .

FROM gcr.io/distroless/static:latest

COPY --from=builder /app/url-redirect .
EXPOSE 8082

CMD ["./url-redirect"]
