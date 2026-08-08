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
# static, not base.
#
# distroless/base ships glibc and OpenSSL so that cgo-linked binaries can run. This
# binary is built CGO_ENABLED=0 and every dependency is pure Go (go-sql-driver/mysql
# included), so none of that was ever loaded — it sat in the image contributing CVEs
# for code that cannot execute. Measured on v1.2.2: 3 findings at high or above, all
# libc6, all marked wont-fix by Debian, so no rebuild would ever have cleared them.
#
# distroless/static carries only ca-certificates, tzdata and /etc/passwd — enough for
# outbound TLS and the nonroot user, with no libc at all. Removing the package beats
# asserting the package is unreachable.
#
# If a future dependency needs cgo, this must go back to base-debian12 and those CVEs
# return as a real exposure rather than a bookkeeping one.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=backend /bin/server /app/server
COPY --from=fe /fe /app/frontend
COPY config.example.yaml /app/config.yaml
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]