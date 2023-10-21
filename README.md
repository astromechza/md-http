# md-http

A simple, but robust, http server that hosts a single Markdown document over http.

I built this primarily for hosting a basic reference page hosting links on my home network and I'm sharing
this because I hope it can fulfill similar use-cases on other internal networks! 

```
Usage: md-http [options...] <filepath>
  -css string
        An optional css file path or url (http:// or https://) to serve in the output
  -debug
        Enable debug logging
  -listen string
        The socket address to listen on (default "0.0.0.0:8080")
  -title string
        The HTML title of the page (default "Landing page")

All options also have an environment variable counterpart: MDHTTP_<option>=<value>
```

## Markdown features

- Most regular common markdown features
- Tables

    | Syntax      | Description |
    | ----------- | ----------- |
    | Header      | Title       |
    | Paragraph   | Text        |

- Footnotes[^1]
- Definition Lists
    
    Term One
    : A definition of Term One

- Automatic header ids: [example](#markdown-features)
- Autolinking: https://github.com

[^1]: The footnote content

## Installation

You can run or install this in a few different ways:

### Compile and run directly with `go run`

Good for testing, but not that useful for deployment into a final environment.

```
$ go run github.com/astromechza/md-http@latest -h
go: downloading github.com/astromechza/md-http v0.0.0-20231021093316-0e979d460e44
Expected a single argument as the markdown filepath!

Usage: md-http [options...] <filepath>
...
```

### Install inside your own Docker image

I don't host a base Docker image for this binary. If you want to embed it in an image, use a multistep builder:

```
FROM golang:1-alpine AS builder
RUN go install -v github.com/astromechza/md-http@latest

FROM alpine
COPY --from=builder /go/bin/md-http /md-http
RUN echo "hello world" > markdown.md
ENTRYPOINT ["/md-http", "markdown.md"]
```
