# Multi-arch multi-stage Dockerfile. GoReleaser will use this to build images via buildx.
# The image expects the binary 'apcupsd_exporter' to be present at the root for the final image

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN apk add --no-cache git ca-certificates && go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags='-s -w' -o /out/apcupsd_exporter ./cmd/apcupsd_exporter

FROM scratch
COPY --from=builder /out/apcupsd_exporter /apcupsd_exporter
ENTRYPOINT ["/apcupsd_exporter"]
