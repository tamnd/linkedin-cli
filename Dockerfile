# Consumed by GoReleaser: it copies the already cross-compiled binary out of the
# build context rather than compiling, so the image build is fast and uses the
# same static binary every other artifact ships.
#
# GoReleaser builds one multi-platform image with buildx and stages each
# platform's binary under a $TARGETPLATFORM directory (e.g. linux/amd64/) in the
# build context, so the COPY line selects the right one through the automatic
# TARGETPLATFORM build arg.
FROM alpine:3.21

ARG TARGETPLATFORM

# ca-certificates for HTTPS to www.linkedins.com; tzdata for sane timestamps.
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -H -u 10001 linkedin \
 && mkdir -p /data \
 && chown linkedin:linkedin /data

COPY $TARGETPLATFORM/linkedin /usr/bin/linkedin

USER linkedin
WORKDIR /data

# All state lives under /data; mount a volume to keep the cache and the store:
#
#   docker run -v ~/data/linkedin:/data ghcr.io/tamnd/linkedin search dune --books
ENV LINKEDIN_DATA_DIR=/data
VOLUME ["/data"]

ENTRYPOINT ["/usr/bin/linkedin"]
