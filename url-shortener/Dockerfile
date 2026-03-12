# ============================================================
# Stage 1 — Builder
# Uses the official Go image to compile the application binary.
# The result is a single, statically-linked executable.
# ============================================================
FROM golang:1.21-alpine AS builder

# Install build tools needed by the SQLite CGO-free driver.
RUN apk add --no-cache git

# Set the working directory inside the container.
WORKDIR /app

# Copy dependency manifests first and download them.
# Docker caches this layer so re-builds skip the download step
# when only source files change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the binary.
# CGO_ENABLED=0 produces a fully static binary (no shared C libs needed).
# -ldflags="-w -s" strips debug symbols to reduce binary size.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o url-shortener .

# ============================================================
# Stage 2 — Runtime
# Uses a minimal Alpine image — only ~5 MB — containing
# just the compiled binary and nothing else.
# ============================================================
FROM alpine:3.19

# Add CA certificates so the app can make HTTPS calls if needed.
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy only the compiled binary from the builder stage.
COPY --from=builder /app/url-shortener .

# The database file will be stored in /app/data so it can be
# mounted as a Docker volume to persist data across container restarts.
RUN mkdir -p /app/data

# Tell the app where to create the SQLite file.
ENV DB_PATH=/app/data/urls.db

# Expose the port the API listens on.
EXPOSE 8080

# Run the binary.
CMD ["./url-shortener"]
