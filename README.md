# gonews
A simple RSS feed aggregator, written in Go

# How-To

Run locally:

```
docker-compose up
```

The web app is then accessible at localhost:8080

# Development

To get a shell in the container, with dev tools available, run:

```
docker-compose -f docker-compose-dev.yml
```

Then run the following to get the container name:

```
docker-compose ps
```

Then run the following to get a shell in the container:

```
docker exec -it <container name> /usr/bin/env bash
```
