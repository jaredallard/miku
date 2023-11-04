# syntax=docker/dockerfile:1

# Hardcoding amd64 for now to allow 90% of systems to build the image.
# We don't really need a platform, except to download ca-certs...
# however, for multi-arch this breaks on systems that don't have QEMU
# setup. So, we just use linux/amd64 since build infra uses that and
# macOS works with that OOTB.
FROM --platform=amd64 debian:bookworm-slim as builder
# hadolint disable=DL3008 # Why: CA certs.
RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update -y && \
  apt-get install --no-install-recommends -y ca-certificates

FROM scratch
ENTRYPOINT ["/usr/local/bin/miku"]

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY miku /usr/local/bin/
