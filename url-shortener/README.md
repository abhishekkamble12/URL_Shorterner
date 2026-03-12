# URL Shortener API

A production-quality URL Shortener REST API built with **Go**, **Gin**, **GORM**, and **SQLite**.

Convert long URLs into compact short links and track how many times each link has been clicked.

---

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation & Running Locally](#installation--running-locally)
- [API Endpoints](#api-endpoints)
- [Example Requests (curl)](#example-requests-curl)
- [Postman Collection](#postman-collection)
- [Docker Usage](#docker-usage)
- [Switching to PostgreSQL](#switching-to-postgresql)
- [Environment Variables](#environment-variables)

---

## Features

| Feature | Endpoint | Method |
|---|---|---|
| API info / landing page | `/` | GET |
| Shorten a URL | `/api/shorten` | POST |
| Redirect to original URL | `/:shortcode` | GET |
| View click analytics | `/api/stats/:shortcode` | GET |
| List all shortened URLs | `/api/urls` | GET |
| Health check | `/api/health` | GET |

---

## Project Structure

```
url-shortener/
│
├── main.go                    # Entry point — starts the server
│
├── config/
│   └── database.go            # Opens the DB connection, runs GORM migrations
│
├── models/
│   └── url.go                 # URL struct — maps to the `urls` database table
│
├── controllers/
│   └── url_controller.go      # HTTP handlers for every endpoint
│
├── routes/
│   └── routes.go              # Wires URL paths → controller functions
│
├── utils/
│   └── generator.go           # Generates unique 7-character short codes
│
├── go.mod                     # Go module definition and dependencies
├── go.sum                     # Dependency checksums (auto-generated)
└── Dockerfile                 # Multi-stage Docker build
```

---

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) *(optional — only needed for containerised runs)*

---

## Installation & Running Locally

```bash
# 1. Clone or navigate to the project
cd url-shortener

# 2. Download all dependencies
go mod tidy

# 3. Start the server (creates urls.db automatically)
go run main.go
```

The server starts at **http://localhost:8080**.

---

## API Endpoints

### POST `/api/shorten` — Create a short URL

**Request body:**
```json
{
  "url": "https://example.com/some/very/long/path?query=value"
}
```

**Success response (201 Created):**
```json
{
  "short_url": "http://localhost:8080/abc1234"
}
```

**Error response (400 Bad Request):**
```json
{
  "error": "Invalid request. Please provide a valid 'url' field."
}
```

---

### GET `/:shortcode` — Redirect to original URL

Performs an **HTTP 302** redirect to the original URL and increments the click counter.

**Not found (404):**
```json
{
  "error": "Short URL not found."
}
```

---

### GET `/api/stats/:shortcode` — View analytics

**Success response (200 OK):**
```json
{
  "original_url": "https://example.com/some/very/long/path",
  "shortcode":    "abc1234",
  "clicks":       42,
  "created_at":   "2026-03-12"
}
```

---

### GET `/api/urls` — List all URLs

**Success response (200 OK):**
```json
{
  "total": 2,
  "urls": [
    {
      "id": 2,
      "original_url": "https://golang.org",
      "short_code":   "xyz9876",
      "clicks":       5,
      "created_at":   "2026-03-12T10:00:00Z"
    },
    {
      "id": 1,
      "original_url": "https://example.com",
      "short_code":   "abc1234",
      "clicks":       42,
      "created_at":   "2026-03-11T08:30:00Z"
    }
  ]
}
```

---

### GET `/api/health` — Liveness probe

```json
{ "status": "ok" }
```

---

## Example Requests (curl)

```bash
# 1. Shorten a URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/golang/go"}'

# Expected:
# { "short_url": "http://localhost:8080/abc1234" }

# -----------------------------------------------------------

# 2. Redirect (follow the redirect with -L)
curl -L http://localhost:8080/abc1234

# The browser (or curl with -L) lands on https://github.com/golang/go

# -----------------------------------------------------------

# 3. Get stats
curl http://localhost:8080/api/stats/abc1234

# Expected:
# { "original_url": "...", "shortcode": "abc1234", "clicks": 1, "created_at": "2026-03-12" }

# -----------------------------------------------------------

# 4. List all URLs
curl http://localhost:8080/api/urls

# -----------------------------------------------------------

# 5. Health check
curl http://localhost:8080/api/health
```

---

## Docker Usage

### Build the image

```bash
docker build -t url-shortener .
```

### Run with a persistent volume

```bash
docker run -p 8080:8080 \
  -v url-shortener-data:/app/data \
  url-shortener
```

The `-v` flag mounts a named volume so the SQLite database survives container restarts.

### Run with a custom base URL

```bash
docker run -p 8080:8080 \
  -e BASE_URL=https://short.example.com \
  -v url-shortener-data:/app/data \
  url-shortener
```

---

## Switching to PostgreSQL

Only **two files** need to change.

### 1. `go.mod` — add the Postgres driver

The `go.mod` already includes `gorm.io/driver/postgres`. After running
`go mod tidy` it will be downloaded automatically.

### 2. `config/database.go` — swap the driver

Replace the SQLite block:

```go
// BEFORE (SQLite)
import "github.com/glebarez/sqlite"

DB, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
```

with:

```go
// AFTER (PostgreSQL)
import "gorm.io/driver/postgres"

dsn := "host=localhost user=postgres password=secret dbname=urlshortener port=5432 sslmode=disable"
DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
```

Everything else — models, controllers, routes — stays exactly the same.

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port the server listens on |
| `DB_PATH` | `urls.db` | Path to the SQLite database file |
| `BASE_URL` | `http://localhost:8080` | Prefix prepended to generated short URLs |
| `GIN_MODE` | `debug` | Set to `release` in production to silence Gin debug logs |
