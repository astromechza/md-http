# md-http

A simple, but robust, http server that hosts a single Markdown document over http.

I built this for hosting a basic reference page hosting links on my home network and I'm sharing
this because I hope it can fulfill similar use-cases on other internal networks! 

- [Installation](#installation)
- [Deployment](#deployment)
- [FAQ](#faq)
- [Markdown features](#markdown-features)

```
Usage: md-http [options...] <filepath>
  -css string
    	An optional css file path or url (http:// or https://) to serve in the output
  -debug
    	Enable debug logging
  -favicon string
    	An optional favicon file path or url (http:// or https://) to serve with the output
  -jsonlog
    	Switch to structured json logging
  -listen string
    	The socket address to listen on (default "0.0.0.0:8080")
  -title string
    	The HTML title of the page (default "Landing page")

All options also have an environment variable counterpart: MDHTTP_<option>=<value>.
More details about this binary can be found at the source repo: https://github.com/astromechza/md-http.
```

The screenshot below was taken when I ran the following: `go run github.com/astromechza/md-http@v1.1.0 -css https://cdn.jsdelivr.net/npm/water.css@2/out/water.css ./README.md`.

![screenshot of rendered README.md](screenshot.jpg)

## Installation

You can run or install this in a few different ways. In all cases, check https://github.com/astromechza/md-http/releases for the latest version tag.

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

I don't host a Docker image for this binary. If you want to embed it in an image, use a multistep builder. See [Dockerfile](./Dockerfile):

```
FROM golang:1-alpine AS builder
RUN go install -v github.com/astromechza/md-http@latest

FROM alpine
COPY --from=builder /go/bin/md-http /md-http
RUN echo "hello world" > markdown.md
ENTRYPOINT ["/md-http", "markdown.md"]
```

### Git clone and build

If the above options do not work, because github.com does not resolve, or dependencies cannot be found,
clone this repo and build with the vendored dependencies.

```
$ git clone https://github.com/astromechza/md-http.git
$ cd md-http
$ go build -o md-http .
$ ./md-http README.md
```

## Deployment

We suggest:

- Build a container with any custom css (`-css example.css`) and favicon (`-favicon example.ico`) embedded.
- Customise the page title using environment variables (`MDHTTP_title`).
- Setup the liveness and readiness checks to point towards the `/healthz` route on the main interface.
- Configure the ingress or proxy to add caching, metrics, tracing, or any other value added extras.

## FAQ

### What if I want to serve a directory of files, not just 1?

Unfortunately, this isn't the project for you. You are probably looking for something more fully featured.

### What if I want to host images as well?

Again, unfortunately that isn't a priority for this project. Host the image elsewhere and embed a link to it.

Alternatively, you can use raw html to embed a svg image (not visible on github):

<svg width="70" height="70">
  <rect x="10" y="10" rx="10" ry="10" width="50" height="50"
  style="fill:red;stroke:black;stroke-width:5;opacity:0.5" />
</svg>

Or a base64 encoded image (not visible on github):

<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAIAAADTED8xAAADMElEQVR4nOzVwQnAIBQFQYXff81RUkQCOyDj1YOPnbXWPmeTRef+/3O/OyBjzh3CD95BfqICMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMK0CMO0TAAD//2Anhf4QtqobAAAAAElFTkSuQmCC" />

### What if I need TLS or authentication?

Put this behind a suitable TLS and auth proxy (Nginx, Apache, Traefik, Envoy, etc..).
Adding this functionality directly to md-http would make it too complex to maintain.

### What if I need extra headers injected for caching or other behaviors?

Again, put this behind a suitable proxy.

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
