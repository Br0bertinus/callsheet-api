# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer if go.mod/go.sum unchanged)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /callsheet-api .

# ---- Runtime stage ----
FROM alpine:latest

# ca-certificates is required for outbound HTTPS calls (e.g. TMDB API)
# tzdata is required for time.LoadLocation (used for Pacific-time daily challenge rollover)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /callsheet-api .
COPY config.yaml .

EXPOSE 8080

ENTRYPOINT ["./callsheet-api", "serve"]
