# syntax=docker/dockerfile:1
FROM ghcr.io/jaredallard/scratch-cacerts:latest
ENTRYPOINT ["/usr/local/bin/miku"]
COPY miku /usr/local/bin/
