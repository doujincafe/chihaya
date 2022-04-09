FROM golang:alpine AS build-env

# Install OS-level dependencies.
RUN apk add --no-cache curl git alpine-sdk

# Install golang tools
RUN go install -v golang.org/x/tools/gopls@latest && \
    go install -v github.com/ramya-rao-a/go-outline@latest && \
    go install -v github.com/go-delve/delve/cmd/dlv@latest && \
    go install -v honnef.co/go/tools/cmd/staticcheck@latest && \
    go install -v github.com/go-delve/delve/cmd/dlv@latest