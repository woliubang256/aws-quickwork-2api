# syntax=docker/dockerfile:1

# ---------- build ----------
# Build on the native builder platform and cross-compile: pure-Go code, so no
# emulation is needed for linux/arm64.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

# The official Go image pins GOTOOLCHAIN=local; allow the patch toolchain that
# go.mod asks for to be fetched instead of failing the build.
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# Cache the module download separately from the source tree.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false \
        -ldflags="-s -w" \
        -o /out/quickproxy ./cmd/quickproxy

# ---------- runtime ----------
FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 10001 -S quickwork \
    && adduser -u 10001 -S -D -G quickwork -h /data -s /sbin/nologin quickwork \
    && mkdir -p /data \
    && chown -R 10001:10001 /data

COPY --from=build /out/quickproxy /usr/local/bin/quickproxy
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY LICENSE /usr/share/licenses/quickproxy/LICENSE

RUN chmod 0755 /usr/local/bin/quickproxy /usr/local/bin/docker-entrypoint.sh

# config.json, quickwork.db and runtime state live on the volume.
WORKDIR /data
USER 10001:10001

ENV QW_CONFIG=/data/config.json

EXPOSE 8080
VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -q -T 3 -O /dev/null http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["serve"]
