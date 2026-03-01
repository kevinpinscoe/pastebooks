# --- build backend ---
FROM golang:1.25 AS backend
WORKDIR /src

# 1) seed deps cache
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 2) bring in the source
COPY backend/ ./

# 3) ensure go.sum exists for all imports
# This was just for quick start development
# and since we are moving to Github Actions
# This step will no longer be needed
# RUN go mod tidy

# 4) build
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server .

# --- build frontend (static) ---
FROM node:22-alpine3.21 AS fe
WORKDIR /fe
COPY frontend/ ./
# (No build step needed for vanilla JS; keep stage for future toolchains)


# --- final image ---
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=backend /bin/server /app/server
COPY --from=fe /fe /app/frontend
COPY config.example.yaml /app/config.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]