![Pucora](pucora.png)

# Pucora

Pucora is an extensible, ultra-high performance API Gateway that helps you effortlessly adopt microservices and secure communications. Pucora is easy to operate and run and scales out without a single point of failure.

**Pucora Community Edition** (or *Pucora-CE*) is an open-source API gateway.

**WebSockets:** RFC-6455 multiplex and direct proxy to `ws://` / `wss://` backends — see [docs/websockets.md](docs/websockets.md) for configuration, JWT on upgrade, and Docker Compose examples.

## Benefits

- **Easy integration** of an ultra-high performance gateway.
- **Effortlessly transition to microservices** and Backend For Frontend implementations.
- **True linear scalability**: Thanks to its **stateless design**, every Pucora node can operate independently in the cluster without any coordination or centralized persistence.
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
- **Multi-layer rate-limiting** for the end-user and between Pucora and your services, including bursting, load balancing, and circuit breaker.
- **Telemetry** and dashboards of all sorts: Datadog, Zipkin, Jaeger, Prometheus, Grafana...
- **WebSockets** (RFC-6455): multiplexed or direct proxy to `ws://` / `wss://` backends — see [docs/websockets.md](docs/websockets.md) and the standalone module [pucora/pucora-websocket](https://github.com/pucora/pucora-websocket)
- **Extensible** with Go plugins, Lua scripts, Martian, or Google CEL spec.

## Run

Build the binary:

```
make build
```

Run with the sample configuration:

```
./pucora run -c pucora.json
```

For WebSocket-only local testing (requires a `ws://` backend on port 8081):

```
./pucora run -c pucora-ws.json
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

Published on Docker Hub: **[pucora/pucora](https://hub.docker.com/r/pucora/pucora)**

```
docker pull pucora/pucora:2.0.0
```

On each GitHub release, CI builds and pushes `pucora/pucora:$TAG` when `DOCKER_USERNAME` and `DOCKER_PASSWORD` repo secrets are set (`DOCKER_USERNAME` = `pucora`).

Build locally:

```
make docker
```

Run it:

```
docker run -it -p "8080:8080" -v $(pwd)/pucora.json:/etc/pucora/pucora.json pucora/pucora:2.0.0 run -c /etc/pucora/pucora.json
```

## Kubernetes / Helm

Deploy to Kubernetes with the official Helm chart from the [pucora-ce](https://github.com/pucora/pucora-ce) repository:

```bash
git clone https://github.com/pucora/pucora-ce.git
cd pucora-ce
helm install my-gateway ./deploy/helm/pucora
```

See [deploy/helm/pucora/README.md](deploy/helm/pucora/README.md) for configuration modes (ConfigMap vs immutable image), Ingress, HPA, PDB, and Prometheus integration.

## Build

See the required Go version in the `Makefile`, and then:

```
make build
```

### Standalone clone (any directory)

`go.mod` uses **GitHub module paths and version tags only** — no `replace ../...` directives. After cloning this repo anywhere:

```
git clone https://github.com/pucora/pucora-ce.git
cd pucora-ce
make build
```

Go downloads dependencies from `github.com/pucora/*` at the versions pinned in `go.mod`.

### Local monorepo (all sibling modules)

When developing several Pucora modules together, use a **workspace file** (not committed to individual repos):

```
# From pucora-ce, with sibling repos checked out under the same parent directory:
./scripts/init-workspace.sh   # writes ../go.work listing local modules
make build                    # uses sibling source instead of published tags
```

Delete `go.work` (or run builds with `GOWORK=off`) to switch back to published GitHub modules.

To publish a updated module tag after local changes:

```
./scripts/publish-fork-module.sh pucora-websocket v2.0.8
```

Or, if you don't have or don't want to install `go`, you can build it using the golang docker container:

```
make build_on_docker
```

## Configuration

Pucora uses a JSON configuration format. Configuration files are named `pucora.json` and placed in `/etc/pucora/`.

### WebSockets

Pucora CE supports RFC-6455 WebSocket proxying (multiplex and direct modes, JWT on upgrade, reconnect, and OTEL metrics). See the full guide:

- [docs/websockets.md](docs/websockets.md) — configuration, envelope protocol, JWT, Docker Compose stack
- [tests/fixtures/ws_*.json](tests/fixtures/) — sample configs
- [pucora-ws.json](pucora-ws.json) — minimal direct-mode sample
- 

## License

Apache 2.0



