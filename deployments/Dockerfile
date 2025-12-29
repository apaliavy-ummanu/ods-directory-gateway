# Server -----------------------------------------------------------------

# Builder stage
FROM golang:1.24.5 AS ods-gateway-server-builder

# Disable CGO so we can use a minimal runtime image
ENV CGO_ENABLED=0
WORKDIR /usr/src/app

# Install build tools (make, git, etc.), if not already present
RUN apt-get update && \
    apt-get install -y --no-install-recommends make ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Go module files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source
COPY . .

# Use a writable cache directory for Go build cache
ENV GOCACHE=/root/.cache/go-build

# Build the server binary; adjust target if needed
RUN make build-server

# Runtime stage
FROM alpine:3.20 AS server

# Add curl (required for healthcheck) and CA certificates
RUN apk --no-cache add curl ca-certificates

# Ensure /go/bin exists
RUN mkdir -p /go/bin /go/config/aws /go/config/local_docker

# Copy the built binary
COPY --from=ods-gateway-server-builder /usr/src/app/bin/api /go/bin/api

# Copy CA certificates from Alpine package (builder copy is not needed)
# but keep this line if your builder image has additional custom CAs:
# COPY --from=ods-gateway-server-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV PATH="/go/bin:${PATH}"

ENTRYPOINT ["/go/bin/api"]