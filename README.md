![Velonetics](velonetics.png)

# Velonetics

Velonetics is an extensible, ultra-high performance API Gateway that helps you effortlessly adopt microservices and secure communications. Velonetics is easy to operate and run and scales out without a single point of failure.

**Velonetics Community Edition** (or *Velonetics-CE*) is an open-source API gateway.

**WebSockets:** RFC-6455 multiplex and direct proxy to `ws://` / `wss://` backends — see [docs/websockets.md](docs/websockets.md) for configuration, JWT on upgrade, and Docker Compose examples.

## Benefits

- **Easy integration** of an ultra-high performance gateway.
- **Effortlessly transition to microservices** and Backend For Frontend implementations.
- **True linear scalability**: Thanks to its **stateless design**, every Velonetics node can operate independently in the cluster without any coordination or centralized persistence.
- **Low operational cost**: +70K reqs/s on a single instance of regular size. Super low memory consumption with high traffic (usually under 50MB w/ +1000 concurrent). Fewer machines. Smaller machines. Lower budget.
- **Platform-agnostic**. Whether you work in a Cloud-native environment (e.g., Kubernetes) or self-hosted on-premises.
- **No vendor lock-in**: Reuse the best existing open-source and proprietary tools rather than having everything in the gateway (telemetry, identity providers, etc.)
- **API Lifecycle**: Using **GitOps** and **declarative configuration**.
- **Decouple clients** from existing services. Create new APIs without changing your existing API contracts.

## Technical features

- **Content aggregation**, composition, and filtering: Create views and mashups of aggregated content from your APIs.
- **Content Manipulation and format transformation**: Change responses, convert transparently from XML to JSON, and vice-versa.
- **Security**: Zero-trust policy, CORS, OAuth, JWT, HSTS, clickjacking protection, HPKP, MIME-Sniffing prevention, XSS protection...
- **Concurrent calls**: Serve content faster than consuming backends directly.
- **SSL** and  **HTTP2** ready
- **Throttling**: Limits of usage in the router and proxy layers
- **Multi-layer rate-limiting** for the end-user and between Velonetics and your services, including bursting, load balancing, and circuit breaker.
- **Telemetry** and dashboards of all sorts: Datadog, Zipkin, Jaeger, Prometheus, Grafana...
- **WebSockets** (RFC-6455): multiplexed or direct proxy to `ws://` / `wss://` backends — see [docs/websockets.md](docs/websockets.md) and the standalone module [velonetics/velonetics-websocket](https://github.com/velonetics/velonetics-websocket)
- **Extensible** with Go plugins, Lua scripts, Martian, or Google CEL spec.

## Run

Build the binary:

```
make build
```

Run with the sample configuration:

```
./velonetics run -c velonetics.json
```

For WebSocket-only local testing (requires a `ws://` backend on port 8081):

```
./velonetics run -c velonetics-ws.json
# or: make ws-compose-test
```

Now see [http://localhost:8080/__health](http://localhost:8080/__health). The gateway is listening.

## Docker

### WebSocket local stack

Gateway + mock backend (direct, multiplex, JWT):

```bash
make ws-compose-test   # build, smoke test, tear down
```

See [examples/websocket/README.md](examples/websocket/README.md).

### Single image

Published on Docker Hub: **[niteesh20/velonetics](https://hub.docker.com/r/niteesh20/velonetics)**

```
docker pull niteesh20/velonetics:2.0.0
```

On each GitHub release, CI builds and pushes `niteesh20/velonetics:$TAG` when `DOCKER_USERNAME` and `DOCKER_PASSWORD` repo secrets are set (`DOCKER_USERNAME` = `niteesh20`).

Build locally:

```
make docker
```

Run it:

```
docker run -it -p "8080:8080" -v $(pwd)/velonetics.json:/etc/velonetics/velonetics.json niteesh20/velonetics:2.0.0 run -c /etc/velonetics/velonetics.json
```

## Build

See the required Go version in the `Makefile`, and then:

```
make build
```

Or, if you don't have or don't want to install `go`, you can build it using the golang docker container:

```
make build_on_docker
```

## Configuration

Velonetics uses a JSON configuration format compatible with Velonetics configs. Configuration files are named `velonetics.json` and placed in `/etc/velonetics/`.

Legacy Velonetics namespace keys in `extra_config` are still accepted for backward compatibility.

### WebSockets

Velonetics CE supports RFC-6455 WebSocket proxying (multiplex and direct modes, JWT on upgrade, reconnect, and OTEL metrics). See the full guide:

- [docs/websockets.md](docs/websockets.md) — configuration, envelope protocol, JWT, Docker Compose stack
- [tests/fixtures/ws_*.json](tests/fixtures/) — sample configs
- [velonetics-ws.json](velonetics-ws.json) — minimal direct-mode sample

## License

Apache 2.0
