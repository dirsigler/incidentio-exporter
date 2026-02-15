FROM cgr.dev/chainguard/busybox:latest

COPY incidentio-exporter /usr/bin/incidentio-exporter

EXPOSE 9193

USER nonroot:nonroot

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:9193/metrics || exit 1

ENTRYPOINT ["/usr/bin/incidentio-exporter"]
