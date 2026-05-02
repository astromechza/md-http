FROM gcr.io/distroless/static-debian12:nonroot
COPY md-http /md-http
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/md-http"]
