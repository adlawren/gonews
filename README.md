# gonews

A simple RSS feed aggregator, written in Go

# Setup

Copy ```config.toml.example``` to ```.config/config.toml```, edit as needed

# How-To

Run locally:

```
docker-compose up
```

The web app is then accessible at localhost:8080

## TLS

To run locally with TLS enabled, edit the cert and key file paths in `docker-compose.tls.yml` accordingly, then run:

```
sudo docker-compose -f docker-compose.yml -f docker-compose.tls.yml up
```

# Development

To get a shell in the container, run:

```
sudo docker-compose -f docker-compose.yml -f docker-compose.dev.yml run --service-ports --rm web bash
```
