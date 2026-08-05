# --- build backend ---
FROM golang:1.26 AS backend
WORKDIR /src

# Include this stage in the SBOM attestation.
#
# The final image is distroless and receives a single statically linked binary, so
# a final-stage-only scan finds the distroless base packages and nothing else — no
# Go module list at all. That SBOM is not empty, which is worse than if it were: it
# looks like a complete, clean inventory. Marking the build stage is what puts the
# actual dependency graph into the attestation.
ARG BUILDKIT_SBOM_SCAN_STAGE=true

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
FROM node:25-alpine3.21 AS fe
WORKDIR /fe

# Same reasoning as the backend stage: only built assets reach the final image, so
# anything this stage introduces is invisible to a final-stage scan.
ARG BUILDKIT_SBOM_SCAN_STAGE=true

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