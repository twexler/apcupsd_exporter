# Multi-arch Dockerfile for GoReleaser dockers_v2.
# GoReleaser builds the binaries and provides them in the build context.
# We just copy the pre-built binary for the target platform.

FROM scratch
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/apcupsd_exporter /apcupsd_exporter
ENTRYPOINT ["/apcupsd_exporter"]
