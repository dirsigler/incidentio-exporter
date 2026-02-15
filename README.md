# Prometheus Exporter for incident.io

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/dirsigler/incidentio-exporter)](https://goreportcard.com/report/github.com/dirsigler/incidentio-exporter)

This is a Prometheus Exporter for the [incident.io](https://incident.io) API. It allows you to monitor your incident.io data in Prometheus and build dashboards in Grafana.

## ⚙️ Metrics

### Incident Metrics

| Name                                       | Labels                                                           | Description                               |
| ------------------------------------------ | ---------------------------------------------------------------- | ----------------------------------------- |
| `incidentio_incidents_total_count`         | -                                                                | Total number of incidents                 |
| `incidentio_incidents_severity_count`      | `severity`                                                       | Number of incidents by severity           |
| `incidentio_incidents_status_count`        | `status`                                                         | Number of incidents by status             |
| `incidentio_incidents_custom_field_count`  | `custom_field_id`, `custom_field_name`, `value`                  | Number of incidents by custom field value |
| `incidentio_incidents_catalog_entry_count` | `catalog_type_id`, `catalog_type_name`, `entry_id`, `entry_name` | Number of incidents by catalog entry      |

### User Metrics

| Name                                 | Labels                 | Description                           |
| ------------------------------------ | ---------------------- | ------------------------------------- |
| `incidentio_users_total_count`       | -                      | Total number of users                 |
| `incidentio_users_base_role_count`   | `role_id`, `role_name` | Number of users by base role          |
| `incidentio_users_custom_role_count` | `role_id`, `role_name` | Number of users with each custom role |

### Exporter Metrics

| Name                                                            | Labels     | Description                                |
| --------------------------------------------------------------- | ---------- | ------------------------------------------ |
| `incidentio_exporter_up`                                        | -          | Whether the exporter is up (1) or down (0) |
| `incidentio_exporter_collection_duration_seconds`               | -          | Duration of last collection in seconds     |
| `incidentio_exporter_api_call_duration_seconds`                 | `endpoint` | Duration of API calls by endpoint          |
| `incidentio_exporter_api_call_errors_total`                     | `endpoint` | Total number of API call errors            |
| `incidentio_exporter_last_collection_success_timestamp_seconds` | -          | Timestamp of last successful collection    |
| `incidentio_exporter_last_collection_attempt_timestamp_seconds` | -          | Timestamp of last collection attempt       |
| `incidentio_exporter_collection_errors_total`                   | -          | Total number of collection errors          |

## 🚀 Deployment

With each [release](https://github.com/dirsigler/incidentio-exporter/releases), a secure-by-default Docker image is available on [GitHub](https://github.com/dirsigler/incidentio-exporter/pkgs/container/incidentio-exporter) and [DockerHub](https://hub.docker.com/repository/docker/dirsigler/incidentio-exporter/general).

### Docker Compose

Here is a sample `docker-compose.yml`:

```yaml
version: "3.8"
services:
  incidentio-exporter:
    image: ghcr.io/dirsigler/incidentio-exporter:latest
    container_name: incidentio-exporter
    restart: unless-stopped
    ports:
      - "9193:9193"
    environment:
      - INCIDENTIO_API_KEY=<YOUR_API_KEY>
```

### Docker

```sh
docker run --rm \
  --interactive --tty \
  --publish 9193:9193 \
  --env INCIDENTIO_API_KEY=<YOUR_API_KEY> \
  ghcr.io/dirsigler/incidentio-exporter:latest
```

## 🚩 Configuration

`$ incidentio-exporter --help`

| Flag                 | Environment Variable          | Description                                          | Default                   |
| -------------------- | ----------------------------- | ---------------------------------------------------- | ------------------------- |
| `--api-key`          | `INCIDENTIO_API_KEY`          | incident.io API key                                  | **required**              |
| `--port`             | `SERVER_PORT`                 | Port to listen on                                    | `9193`                    |
| `--log-level`        | `LOG_LEVEL`                   | Log level (debug, info, warn, error)                 | `info`                    |
| `--api-url`          | `INCIDENTIO_API_URL`          | incident.io API URL                                  | `https://api.incident.io` |
| `--custom-field-ids` | `INCIDENTIO_CUSTOM_FIELD_IDS` | Custom field IDs to track (specified multiple times) | -                         |
| `--catalog-type-ids` | `INCIDENTIO_CATALOG_TYPE_IDS` | Catalog type IDs to track (specified multiple times) | -                         |

### Examples

**Basic usage:**

```sh
incidentio-exporter \
  --api-key=your-api-key
```

**With custom fields and catalog types:**

```sh
incidentio-exporter \
  --api-key=your-api-key \
  --custom-field-ids=field-1 \
  --custom-field-ids=field-2 \
  --catalog-type-ids=type-1
```

**Or using environment variables:**

```sh
export INCIDENTIO_API_KEY=your-api-key
export INCIDENTIO_CUSTOM_FIELD_IDS=field-1,field-2
export INCIDENTIO_CATALOG_TYPE_IDS=type-1

incidentio-exporter
```

## 📝 License

Built with ☕️ and licensed under the [Apache 2.0 License](./LICENSE).
