FROM alpine:3.18
ENTRYPOINT ["/usr/local/bin/miku"]
COPY miku /usr/local/bin/
