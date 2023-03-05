# a23n

A simple authentication and authorization gRPC service.

## Running in Docker

```shell
docker run --rm -it \
  -e A23N_DB_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
  -e A23N_API_SECRET=$(for N in {1..12}; do echo -n $RANDOM; done | base64 | head -c 64) \
  ashep/a23n:latest
```

## Running in Kubernetes

To do.

## Configuring via environment variables

- *required* **string** `A23N_DB_DSN`. PostgreSQL DSN.
- *required* **string** `A23N_API_SECRET`. Encryption key. A random string of length >=32 chars.
- *optional* **int** `A23N_API_TOKEN_TTL`. Authentication token time to live. Default is `86400`.
- *optional* **int** `A23N_SERVER_ADDRESS`. HTTP/gRPC listen address. Default is `localhost:9000`.
- *optional* **int** `A23N_DEBUG`. Set to `1` to enable debug mode. Any other value being ignored.

## gRPC API

To do.

## HTTP API

To do.

## To Do

- Complete this readme.
- Write unit tests.
- Write functional tests.
- Add entities caching.

## Changelog

**0.0.1** (2023-03-05)

Initial version.

## Authors

- [Oleksandr Shepetko](https://shepetko.com). Initial work.
