# Contributing

## Set up your machine

`incidentio-exporter` is written in [Go](https://golang.org/).

Prerequisites:

- [Go 1.22+](https://go.dev/doc/install)

Clone `incidentio-exporter` anywhere:

```sh
git clone git@github.com:dirsigler/incidentio-exporter.git
```

## Test your change

Add your changes to the code and build the binary as well as Dockerimage locally.
You can use the example Docker Compose stack in [./examples/docker/](./examples/docker/) to see the changes live.

Else verify that a `curl http://localhost:9193/metrics` responds with the desired Prometheus metrics.

## Create a commit

Commit messages should be well formatted.

## Submit a pull request

Push your branch to your `incidentio-exporter` fork and open a pull request against the main branch.
