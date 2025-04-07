# Ouroboros

### Development

`docker build -t ouroboros .`

`docker run --rm -p 8080:8080 -p 8081:8081 ouroboros`


The dockerized client exposes a simple (self documented) HTTP API, which can be
reached at

```
http://localhost:8081/docs/
```

With this API, you can get statistics and basic management for
the client.

**Setting Client name**


`docker run --rm -p 8080:8080 -p 8081:8081 -e NAME=kaya ouroboros`
