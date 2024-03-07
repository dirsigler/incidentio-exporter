# Prometheus Exporter for [incident.io](https://incident.io)

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This is an custom Prometheus Exporter for the awesome https://incident.io incident management solution.

Allows your business to monitor all types of pre-configured incidents.
Improve your incident management workflow and better understand the usage of your [incident.io](https://incident.io) installation.

## ⚙️ Metrics

The Incident.io Prometheus Exporter supports all basic pre-configured types of incidents available in [incident.io](https://incident.io).

| Name                                  | Label    | Description                                               |
|---------------------------------------|----------|-----------------------------------------------------------|
| `incidentio_incidents_total_count`    |          | Total count of incidents for the incident.io installation |
| `incidentio_incidents_severity_count` | severity | Total count of incidents labeled per severity             |
| `incidentio_incidents_status_count`   | status   | Total count of incidents labeled per status               |

## 🚀 Deployment

> IMPORTANT: You have to provide the "INCIDENTIO_API_KEY="<MY_API_KEY>" environment variable to your deployment for the Incident.io Prometheus Exporter to work.

---

With each [release](https://github.com/dirsigler/incidentio-exporter/releases) I also provide a [secure by default](https://www.chainguard.dev/chainguard-images) Docker Image.

You can chose from:
- The Image on GitHub => [Incidentio-Exporter on GitHub](https://github.com/dirsigler/incidentio-exporter/pkgs/container/incidentio-exporter)
- The Image on DockerHub => [Incidentio-Exporter on DockerHub](https://hub.docker.com/repository/docker/dirsigler/incidentio-exporter/general)

### Docker
```sh
docker run --rm \
--interactive --tty \
--env INCIDENTIO_API_KEY="<MY_API_KEY>" \
dirsigler/incidentio-exporter:latest
```

You can also enable a logger with Debug mode via the `--log.level=DEBUG` flag.
See the available [configuration](#🚩-configuration)


## 🚩 Configuration

```sh
$ incidentio-exporter --help
Usage of incidentio-exporter:
  -log.level value
    	Configured Log level. (default INFO)
  -server.addr int
    	Address to listen on for HTTP requests. (default 9193)
```


## 📝 License

Built with ☕️ and licensed via [Apache 2.0](./LICENSE)
