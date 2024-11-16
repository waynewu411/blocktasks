ARG GO_VERSION=1.23.0
ARG RUNNER=gcr.io/distroless/base-debian12:nonroot

FROM golang:${GO_VERSION} AS builder
ARG TARGETARCH=amd64
ARG VERSION=latest
ARG BUILD
WORKDIR /build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    make all-mod-download
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOARCH=${TARGETARCH} make build-blocktasks VERSION=${VERSION} BUILD=${BUILD}

FROM ${RUNNER}
    WORKDIR /app
COPY --from=builder --chown=65532:65532 /build/cmd/app /app/app
USER 65532:65532
ENTRYPOINT [ "/app/app" ]
