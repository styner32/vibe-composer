# ---- Build Stage ----
FROM --platform=linux/amd64 golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/server ./cmd/server

# Build migrate binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/migrate ./cmd/migrate

# ---- Runtime Stage ----
FROM --platform=linux/amd64 alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/server /usr/local/bin/server
COPY --from=builder /bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/web /web

WORKDIR /

EXPOSE 8080

CMD ["server"]
