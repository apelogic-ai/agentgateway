FROM cgr.dev/chainguard/static@sha256:60582b2ae6074f641094af0f370d4ab241aab271858a66223dcde7eee9f51638

ARG GOARCH=amd64

COPY testbox-linux-$GOARCH /usr/local/bin/testbox

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/testbox"]
