# callsheet-api

A Go REST API that powers a **Callsheet** game — a *Six Degrees of Kevin Bacon*‑style challenge where players chain actors together through shared movie credits. The server is backed by [The Movie Database (TMDB)](https://www.themoviedb.org/) and caches credit lookups in memory to keep round‑trips fast.

---

## Table of Contents

- [How the Game Works](#how-the-game-works)
- [Requirements](#requirements)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Running the Server](#running-the-server)
- [Running with Docker](#running-with-docker)
- [API Reference](#api-reference)
  - [Search People](#get-searchpeople)
  - [Get Person](#get-peopleid)
  - [Validate Step](#post-gamevalidate-step)
- [Error Responses](#error-responses)
- [Running Tests](#running-tests)
- [Project Structure](#project-structure)

---

## How the Game Works

Players are given **two seed actors** — a starting actor and a target actor — and must build the shortest possible chain connecting them. Each link in the chain must be two actors who appeared in a movie together. Rules:

1. Any actor can only appear **once** in a chain — no revisiting.
2. A step is **valid** only if the current actor and the proposed next actor share at least one movie credit.
3. The API returns the connecting movies so the client can display proof of each link.
4. The goal is to reach the target actor in as **few moves as possible**.

---

## Requirements

- **Go 1.26+** (uses the enhanced `net/http` ServeMux with path‑value extraction)
- A free [TMDB API key](https://developer.themoviedb.org/docs/getting-started)

---

## Getting Started

```bash
git clone https://github.com/Br0bertinus/callsheet-api.git
cd callsheet-api
go mod download
```

Copy the example environment file and add your key:

```bash
cp .env.example .env
# edit .env and set TMDB_API_KEY=<your key>
```

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TMDB_API_KEY` | Yes | API key obtained from TMDB developer portal |

---

## Running the Server

```bash
export TMDB_API_KEY=your_key_here
go run . serve
```

The server starts on **`:8080`** by default. To use a different address:

```bash
go run . serve --addr :9090
```

---

## Running with Docker

### Prerequisites

[Docker Desktop](https://www.docker.com/products/docker-desktop/) (free for personal use) must be installed and running.

### Build the image

```bash
docker build -t callsheet-api .
```

### Run the container

```bash
docker run -p 8080:8080 -e TMDB_API_KEY=your_key_here callsheet-api
```

Or use a `.env` file to avoid passing the key inline:

```bash
docker run -p 8080:8080 --env-file .env callsheet-api
```

The API is available at `http://localhost:8080` once the container starts. Stop it with `Ctrl+C`.

### Useful commands

```bash
# List running containers
docker ps

# Tail logs from a running container
docker logs -f <container-id>

# Stop a container
docker stop <container-id>

# Remove the image
docker rmi callsheet-api
```

### Request logs

Every request is logged to stdout as structured JSON (via [zap](https://github.com/uber-go/zap)) with the method, path, response status, and elapsed time:

```json
{"level":"info","ts":1740398096.123,"msg":"request","method":"GET","path":"/search/people","status":200,"duration":"3.412ms"}
{"level":"info","ts":1740398097.456,"msg":"request","method":"POST","path":"/game/validate-step","status":400,"duration":"81µs"}
```

These are visible in the terminal or in the **Containers** tab of Docker Desktop.

---

## API Reference

All responses use `Content-Type: application/json`.

---

### `GET /health`

Returns the current health status of the server. Useful for container health checks and load balancer probes.

**Example Request**

```bash
curl "http://localhost:8080/health"
```

**Example Response** `200 OK`

```json
{
  "status": "ok"
}
```

---

### `GET /search/people`

Search TMDB for actors or crew members by name.

**Query Parameters**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `q` | string | Yes | Name or partial name to search for |

**Example Request**

```bash
curl "http://localhost:8080/search/people?q=Tom+Hanks"
```

**Example Response** `200 OK`

```json
[
  {
    "id": 31,
    "name": "Tom Hanks",
    "profilePath": "/xndWFsBlClOJFRdhSt4NBwiPq2o.jpg"
  },
  {
    "id": 2227929,
    "name": "Tom Hanks Jr.",
    "profilePath": ""
  }
]
```

---

### `GET /people/{id}`

Fetch a single person by their TMDB person ID.

**Path Parameters**

| Parameter | Type | Description |
|---|---|---|
| `id` | integer | TMDB person ID (must be a positive integer) |

**Example Request**

```bash
curl "http://localhost:8080/people/31"
```

**Example Response** `200 OK`

```json
{
  "id": 31,
  "name": "Tom Hanks",
  "profilePath": "/xndWFsBlClOJFRdhSt4NBwiPq2o.jpg"
}
```

---

### `POST /game/validate-step`

Validate a proposed move in the game. Checks that:

- `nextActorId` has not already been visited in this chain.
- `currentActorId` and `nextActorId` share at least one movie credit.

When the step is valid the response includes all connecting movies as proof.

**Request Body**

| Field | Type | Required | Description |
|---|---|---|---|
| `currentActorId` | integer | Yes | TMDB ID of the actor the player is moving **from** |
| `nextActorId` | integer | Yes | TMDB ID of the actor the player wants to move **to** |
| `visitedActorIds` | integer[] | No | All actor IDs already used in the current chain (prevents revisiting) |

**Example Request — Valid Step**

```bash
curl -X POST "http://localhost:8080/game/validate-step" \
  -H "Content-Type: application/json" \
  -d '{
    "currentActorId": 31,
    "nextActorId": 287,
    "visitedActorIds": [31]
  }'
```

**Example Response** `200 OK` — step is valid

```json
{
  "valid": true,
  "connectingMovies": [
    {
      "id": 13,
      "title": "Forrest Gump",
      "year": 1994
    },
    {
      "id": 857,
      "title": "Saving Private Ryan",
      "year": 1998
    }
  ]
}
```

**Example Request — Invalid Step (already visited)**

```bash
curl -X POST "http://localhost:8080/game/validate-step" \
  -H "Content-Type: application/json" \
  -d '{
    "currentActorId": 31,
    "nextActorId": 287,
    "visitedActorIds": [31, 287]
  }'
```

**Example Response** `200 OK` — step is invalid (actor already in chain)

```json
{
  "valid": false,
  "connectingMovies": []
}
```

**Example Response** `200 OK` — step is invalid (no shared movies)

```json
{
  "valid": false,
  "connectingMovies": []
}
```

---

## Error Responses

All errors return a JSON body with a single `error` field.

```json
{
  "error": "description of the problem"
}
```

| Status | Meaning |
|---|---|
| `400 Bad Request` | Missing or malformed request parameters / body |
| `502 Bad Gateway` | TMDB upstream request failed |

---

## Running Tests

```bash
go test ./...
```

The service layer uses an in‑memory fake TMDB client so tests run without a network connection or API key.

---

## Project Structure

```
callsheet-api/
├── main.go                  # Entry point — calls cmd.Execute()
├── cmd/
│   ├── root.go              # Root Cobra command and Execute()
│   └── serve.go             # `serve` subcommand — initializes dependencies and starts server
├── server/
│   ├── server.go            # Server struct, routes, graceful shutdown
│   ├── health.go            # GET /health
│   ├── people.go            # GET /search/people, GET /people/{id}
│   ├── game.go              # POST /game/validate-step
│   ├── middleware.go        # Logging middleware (method, path, status, latency)
│   └── respond.go           # JSON/error response helpers
├── internal/
│   ├── cache/
│   │   ├── cache.go         # Cache interface
│   │   └── memory.go        # In-memory cache with TTL (default 30 min)
│   ├── client/
│   │   ├── tmdb.go          # TMDBClient interface
│   │   └── tmdb_http.go     # HTTP implementation backed by TMDB API v3
│   ├── domain/
│   │   └── models.go        # Shared domain types (Actor, Movie, request/response structs)
│   └── service/
│       ├── game.go          # Business logic: search, lookup, step validation
│       └── game_test.go     # Unit tests for game service
├── Dockerfile               # Multi-stage build → minimal alpine runtime image
├── .dockerignore            # Excludes .git, .env files, and tests from build context
└── go.mod
```
