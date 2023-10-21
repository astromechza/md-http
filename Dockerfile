FROM golang:1-alpine AS builder
RUN go install -v github.com/astromechza/md-http@latest

FROM alpine
COPY --from=builder /go/bin/md-http /md-http
RUN echo "hello world" > markdown.md
ENTRYPOINT ["/md-http", "markdown.md"]
