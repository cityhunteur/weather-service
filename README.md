# weather-service

 The weather service exposes an endpoint to query the weather forecast for given cities.

## Requirements

* Docker >= 24
* Docker Compose >= 2.19.0
* Go >= 1.20

## Build

```shell
make build
```

## Test

```shell
make test
```

## Run

```shell
make run
```

## API

### Get weather forecasts

```shell
curl --request GET \
  --url http://localhost:8080/v1/weather?city=los%20angels,new%20york,chicago 
```

## TODOs

- [ ] Improve reliability of third-party API clients, e.g. retries, back-off
- [ ] Add tests for API clients using fake
- [ ] Add configurable defaults, e.g. env vars
- [ ] Refactor main handler to perform tasks concurrently if needed
- [ ] Improve accuracy of place search, e.g. using structured query
- [ ] Invalidate cache, evict expired entries and limit cache size

