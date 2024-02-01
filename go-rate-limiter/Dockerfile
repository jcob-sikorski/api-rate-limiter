# Pin specific version for stability
# Use separate stage for building image
# Use debian for easier build utilities
FROM golang:1.22rc2-bullseye AS build-base

WORKDIR /app 

# Copy only files required to install dependencies (better layer caching)
COPY go.mod go.sum ./

# Use cache mount to speed up install of existing dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download

FROM build-base AS build-production

# Add non root user
RUN useradd -u 1001 nonroot

COPY . .

# Compile application during build rather than at runtime
# Add flags to statically link binary
RUN go build \
  -ldflags="-linkmode external -extldflags -static" \
  -tags netgo \
  -o middleware-golang

# Use separate stage for deployable image
FROM scratch

WORKDIR /

# Copy the /etc/passwd file from the build-production stage to the scratch image
COPY --from=build-production /etc/passwd /etc/passwd

# Copy the app binary from the build stage
COPY --from=build-production /app/middleware-golang middleware-golang

COPY --from=build-production /app/config.yaml config.yaml

# Use nonroot user
USER nonroot

# Indicate expected port
EXPOSE 8080

CMD ["/middleware-golang"]
