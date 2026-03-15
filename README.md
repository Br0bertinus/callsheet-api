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
  - [Start Game](#post-game)
  - [Daily Challenge](#get-gamedaily)
  - [Search People](#get-searchpeople)
  - [Search Movies](#get-searchmovies)
  - [Get Person](#get-peopleid)
  - [Validate Step](#post-gamevalidate-step)
- [Error Responses](#error-responses)
- [Running Tests](#running-tests)
- [Daily Challenge](#daily-challenge)
- [Project Structure](#project-structure)

---

## How the Game Works

Players are given **two seed actors** — a starting actor and a target actor — and must build the shortest possible chain connecting them. Each link in the chain requires naming **both** the next actor and the specific movie that connects them. Rules:

1. Any actor can only appear **once** in a chain — no revisiting.
2. Any movie can only be used **once** in a chain — no reusing the same film.
3. A step is **valid** only if the named movie is one that both `currentActorId` and `nextActorId` actually appeared in together.
4. The API returns all shared movies between the two actors so the client can display them as hints or confirm the correct answer.
5. The goal is to reach the target actor in as **few moves as possible**.

For example, to connect **Tom Hanks** to **Ed Norton** a valid chain might be:

> Tom Hanks → *Saving Private Ryan* → Matt Damon → *Ocean's Eleven* → Brad Pitt → *Fight Club* → Ed Norton

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
| `CORS_ORIGIN` | No | Allowed CORS origin (e.g. `https://myapp.com`). Defaults to `*` when unset — fine for local dev, set explicitly in production |
| `DAILY_CHALLENGE_OVERRIDE` | No | Hot-override today's daily challenge pair without redeploying. Format: `<startActorId>,<targetActorId>` (e.g. `500,287`). Takes precedence over all other selection logic for as long as the variable is set. |

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

### `GET /game/daily`

Returns today's fixed start/target actor pair. Every caller on the same UTC calendar day receives the **same two actors**, so all players are solving the identical challenge. The pair is derived deterministically from the date — no randomness per request — and rotates automatically at **00:00 UTC**.

No authentication required. No request body.

**Example Request**

```bash
curl "http://localhost:8080/game/daily"
```

**Example Response** `200 OK`

```json
{
  "startActor": {
    "id": 500,
    "name": "Tom Cruise",
    "profilePath": "/aoksp9ovD4TTxge1SFnWBHhet7a.jpg",
    "popularity": 67.3
  },
  "targetActor": {
    "id": 6193,
    "name": "Leonardo DiCaprio",
    "profilePath": "/wo2hJpn04vbtmh0B9utCFdsQhxM.jpg",
    "popularity": 88.5
  }
}
```

> The response shape is identical to `POST /game` — the frontend can hand the two actors directly to the same gameplay flow.

---

### `POST /game`

Start a new game session. The player uses search to find their two chosen actors, then calls this endpoint to confirm both exist in TMDB and get their full details to bootstrap the client's game state.

**Request Body**

| Field | Type | Required | Description |
|---|---|---|---|
| `startActorId` | integer | Yes | TMDB ID of the actor the player wants to start from |
| `targetActorId` | integer | Yes | TMDB ID of the actor the player wants to reach |

**Example Request**

```bash
curl -X POST "http://localhost:8080/game" \
  -H "Content-Type: application/json" \
  -d '{"startActorId": 31, "targetActorId": 819}'
```

**Example Response** `201 Created`

```json
{
  "startActor": {
    "id": 31,
    "name": "Tom Hanks",
    "profilePath": "/xndWFsBlClOJFRdhSt4NBwiPq2o.jpg"
  },
  "targetActor": {
    "id": 819,
    "name": "Edward Norton",
    "profilePath": "/8nytsqL59SFJTVYVrN72k6qkGgJ.jpg"
  }
}
```

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

### `GET /search/movies`

Search TMDB for movies by title. Use this to look up a movie's ID before submitting it as a chain link.

**Query Parameters**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `q` | string | Yes | Title or partial title to search for |

**Example Request**

```bash
curl "http://localhost:8080/search/movies?q=Saving+Private+Ryan"
```

**Example Response** `200 OK`

```json
[
  {
    "id": 857,
    "title": "Saving Private Ryan",
    "year": 1998
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

Validate a proposed move in the game. The player must name **both** the next actor and the specific movie that connects them. The step is rejected when:

- `nextActorId` has already been visited in this chain.
- `movieId` has already been used in this chain.
- The named movie is not one that both actors actually appeared in together.

The response always includes all valid shared movies between the two actors so the client can surface them as hints or confirm the answer.

**Request Body**

| Field | Type | Required | Description |
|---|---|---|---|
| `currentActorId` | integer | Yes | TMDB ID of the actor the player is moving **from** |
| `nextActorId` | integer | Yes | TMDB ID of the actor the player wants to move **to** |
| `movieId` | integer | Yes | TMDB ID of the movie the player claims connects the two actors |
| `visitedActorIds` | integer[] | No | All actor IDs already used in the current chain (prevents revisiting) |
| `visitedMovieIds` | integer[] | No | All movie IDs already used in the current chain (prevents reuse) |

**Example Request — Valid Step**

```bash
curl -X POST "http://localhost:8080/game/validate-step" \
  -H "Content-Type: application/json" \
  -d '{
    "currentActorId": 31,
    "nextActorId": 287,
    "movieId": 857,
    "visitedActorIds": [31],
    "visitedMovieIds": []
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

**Example Request — Invalid Step (wrong movie named)**

```bash
curl -X POST "http://localhost:8080/game/validate-step" \
  -H "Content-Type: application/json" \
  -d '{
    "currentActorId": 31,
    "nextActorId": 287,
    "movieId": 999,
    "visitedActorIds": [31],
    "visitedMovieIds": []
  }'
```

**Example Response** `200 OK` — step is invalid; `connectingMovies` shows what would have been valid

```json
{
  "valid": false,
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

**Example Response** `200 OK` — step is invalid (actor already in chain or movie already used)

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

## Daily Challenge

### How the pair is selected

On each request to `GET /game/daily` the server applies the following priority order:

1. **Env var override** — set `DAILY_CHALLENGE_OVERRIDE=<startId>,<targetId>` to immediately force a specific pair for all users without redeploying. Unset it when the day is over.
2. **Code-level override map** — add an entry to `internal/service/daily_overrides.go` for planned editorial picks (e.g. Oscar night, a film anniversary). Commit and deploy ahead of time; stale entries are ignored automatically.
3. **Seeded PRNG** — the UTC date string (`"2026-03-14"`) is hashed with FNV-64a to seed a local `rand`, which picks two distinct actors from the pool. Same date always yields the same pair.

### Actor pool

The pool of eligible actors lives in `internal/service/actor_pool.go` and is generated by a script — **do not edit it by hand**.

To add or remove actors, edit the names list in `scripts/rebuild_actor_pool/main.go` then run:

```bash
TMDB_API_KEY=<key> go run ./scripts/rebuild_actor_pool
```

The script searches TMDB for each name, picks the highest-popularity result, and rewrites `actor_pool.go` with verified IDs.

### Verifying the pool

To confirm every ID in the pool still maps to the expected actor name:

```bash
TMDB_API_KEY=<key> go run ./scripts/verify_actor_pool
```

Exits `0` when all IDs match, `1` with a mismatch/not-found report otherwise — CI-friendly.

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
│   ├── movies.go            # GET /search/movies
│   ├── game.go              # POST /game, GET /game/daily, POST /game/validate-step
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
│       ├── game.go          # Business logic: search, lookup, step validation, daily challenge
│       ├── actor_pool.go    # Generated actor pool for daily challenge (run rebuild_actor_pool to update)
│       ├── daily_overrides.go  # Date-keyed editorial overrides for the daily challenge
│       └── game_test.go     # Unit tests for game service
├── scripts/
│   ├── rebuild_actor_pool/
│   │   └── main.go          # Searches TMDB by name and regenerates actor_pool.go
│   └── verify_actor_pool/
│       └── main.go          # Checks every ID in actor_pool.go against the TMDB API
├── Dockerfile               # Multi-stage build → minimal alpine runtime image
├── .dockerignore            # Excludes .git, .env files, and tests from build context
└── go.mod
```
