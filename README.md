# gonews
A simple RSS feed aggregator, written in Go

# How-To

Run locally:

```
docker build -t gonews .
docker run -p 8080:8080 -it gonews
```

The web app is then accessible at localhost:8080

# Development

To build/run from a local application directory:

```
docker run -e GONEWS_ENV=DEV -p 8080:8080 -it -v /local/path/to/gonews:/go/src/gonews gonews
```
