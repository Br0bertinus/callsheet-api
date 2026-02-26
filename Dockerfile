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
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /callsheet-api .

EXPOSE 8080

ENTRYPOINT ["./callsheet-api"]
