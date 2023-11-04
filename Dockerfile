# syntax=docker/dockerfile:1
FROM debian:bookworm-slim as builder
# hadolint disable=DL3008 # Why: CA certs.
RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update -y && \
  apt-get install --no-install-recommends -y ca-certificates

FROM scratch
ENTRYPOINT ["/usr/local/bin/miku"]

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY miku /usr/local/bin/
