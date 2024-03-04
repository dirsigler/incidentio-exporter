FROM cgr.dev/chainguard/static:latest

ARG UID=65532
ARG GID=65532

COPY --chown=${UID}:${GID} incidentio-exporter /usr/bin/incidentio-exporter

EXPOSE 9193

USER nonroot

ENTRYPOINT [ "/usr/bin/incidentio-exporter" ]
