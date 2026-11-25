# syntax=docker/dockerfile:1
FROM gcr.io/distroless/static-debian13:nonroot
ENTRYPOINT ["/usr/local/bin/miku"]

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/miku /usr/local/bin/
