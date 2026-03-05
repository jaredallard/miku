# syntax=docker/dockerfile:1
FROM ghcr.io/jaredallard/scratch-cacerts:latest
ENTRYPOINT ["/usr/local/bin/miku"]

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/miku /usr/local/bin/
