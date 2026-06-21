.PHONY: all build test jwk-aggregator-plugin test-auth check-fixtures-auth check-fixtures-endpoints test-endpoints

# This Makefile is a simple example that demonstrates usual steps to build a binary that can be run in the same
# architecture that was compiled in. The "ldflags" in the build assure that any needed dependency is included in the
# binary and no external dependencies are needed to run the service.

BIN_NAME :=pucora
OS := $(shell uname | tr '[:upper:]' '[:lower:]')
MODULE := github.com/pucora/pucora-ce/v2
WORKSPACE_ROOT := $(abspath ..)
ifneq ($(wildcard $(WORKSPACE_ROOT)/go.work),)
export GOWORK := $(WORKSPACE_ROOT)/go.work
export GOSUMDB := off
endif
VERSION := 2.3.2
SCHEMA_VERSION := 2.13
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo "unknown")
PKGNAME := pucora
LICENSE := Apache 2.0
VENDOR=
URL := https://pucora.io
RELEASE := 0
USER := pucora
ARCH := amd64
DESC := High performance API gateway. Aggregate, filter, manipulate and add middlewares
MAINTAINER := Pucora Team <community@pucora.io>
DOCKER_USER := pucora
DOCKER_WDIR := /tmp/fpm
DOCKER_FPM := $(DOCKER_USER)/fpm
DOCKER_BUILDER := $(DOCKER_USER)/builder
DOCKER_CE := $(DOCKER_USER)/pucora
GOLANG_VERSION := 1.26.2
GLIBC_VERSION := $(shell sh find_glibc.sh)
ALPINE_VERSION := 3.23
OS_TAG :=
EXTRA_LDFLAGS :=

FPM_OPTS=-s dir -v $(VERSION) -n $(PKGNAME) \
  --license "$(LICENSE)" \
  --vendor "$(VENDOR)" \
  --maintainer "$(MAINTAINER)" \
  --architecture $(ARCH) \
  --url "$(URL)" \
  --description  "$(DESC)" \
	--config-files etc/ \
  --verbose

DEB_OPTS= -t deb --deb-user $(USER) \
	--depends ca-certificates \
	--depends rsyslog \
	--depends logrotate \
	--before-remove builder/scripts/prerm.deb \
  --after-remove builder/scripts/postrm.deb \
	--before-install builder/scripts/preinst.deb

RPM_OPTS =--rpm-user $(USER) \
	--depends rsyslog \
	--depends logrotate \
	--before-install builder/scripts/preinst.rpm \
	--before-remove builder/scripts/prerm.rpm \
  --after-remove builder/scripts/postrm.rpm

all: test test-websocket test-graphql test-streaming test-soap test-grpc test-pubsub

build: cmd/pucora-ce/schema/schema.json
	@echo "Building the binary..."
ifndef GOWORK
	@go get .
endif
	@go build -ldflags="-X ${MODULE}/pkg.Version=${VERSION} -X github.com/pucora/lura/v2/core.PucoraVersion=${VERSION} \
	-X github.com/pucora/lura/v2/core.GlibcVersion=${GLIBC_VERSION} ${EXTRA_LDFLAGS}" \
	-o ${BIN_NAME} ./cmd/pucora-ce
	@echo "You can now use ./${BIN_NAME}"

test: build
	go test -v ./tests

test-websocket:
	cd ../pucora-websocket && go test ./...

test-streaming: build
	cd ../lura && go test ./config -run Streaming -count=1
	cd ../lura && go test ./proxy -run 'TestIsStreamingEndpoint|TestStreamCopy|TestNopHTTPResponseParser' -count=1
	cd ../lura && go test ./router/gin -run 'TestRender_noop' -count=1
	cd ../pucora-audit && go test ./... -run 'Test_hasStreaming' -count=1
	go test ./tests -run 'TestStreaming' -count=1 -v

test-soap:
	cd ../pucora-soap && go test ./...

test-grpc:
	cd ../pucora-grpc && go test ./...

test-pubsub:
	cd ../pucora-pubsub && go test ./...

test-graphql:
	cd ../pucora-lura && go test ./transport/http/client/graphql/... ./proxy/... -run 'GraphQL|GetOptions|Resolve|ExtraConfig' -count=1

check-fixtures: build
	./${BIN_NAME} check -c tests/fixtures/ws_direct.json
	./${BIN_NAME} check -c tests/fixtures/ws_multiplex.json
	./${BIN_NAME} check -c tests/fixtures/ws_jwt.json
	./${BIN_NAME} check -c pucora-ws.json
	./${BIN_NAME} check -c tests/fixtures/soap_country.json
	./${BIN_NAME} check -c tests/fixtures/grpc_client.json
	./${BIN_NAME} check -c tests/fixtures/grpc_client_mapping.json
	./${BIN_NAME} check -c tests/fixtures/grpc_server.json
	./${BIN_NAME} check -c tests/fixtures/grpc_server_mixed.json
	./${BIN_NAME} check -c tests/fixtures/grpc_server_jwt.json
	./${BIN_NAME} check -c tests/fixtures/sse_stream.json
	./${BIN_NAME} check -c tests/fixtures/pubsub_mem.json

check-fixtures-graphql: build
	./${BIN_NAME} check -c tests/fixtures/graphql_mutation_post.json
	./${BIN_NAME} check -c tests/fixtures/graphql_query_get.json
	./${BIN_NAME} check -c tests/fixtures/graphql_query_path.json
	./${BIN_NAME} check -c tests/fixtures/graphql_proxy.json
	./${BIN_NAME} check -c tests/fixtures/graphql_proxy_cors.json
	./${BIN_NAME} check -c tests/fixtures/graphql_federation.json
	./${BIN_NAME} check -c tests/fixtures/graphql_jwt.json
	./${BIN_NAME} check -c tests/fixtures/graphql_ratelimit.json

check-fixtures-auth: build
	./${BIN_NAME} check -c tests/fixtures/api_keys.json
	./${BIN_NAME} check -c tests/fixtures/basic_auth.json
	./${BIN_NAME} check -c tests/fixtures/backend_gcp.json
	./${BIN_NAME} check -c tests/fixtures/backend_sigv4.json
	./${BIN_NAME} check -c tests/fixtures/backend_ntlm.json
	./${BIN_NAME} check -c tests/fixtures/revoke_server.json
	./${BIN_NAME} check -c tests/fixtures/multi_idp.json
	./${BIN_NAME} check -c tests/fixtures/backend_client_tls.json
	./${BIN_NAME} check -c tests/fixtures/prod_auth_baseline.json

audit-prod: build
	./${BIN_NAME} audit -c tests/fixtures/prod_auth_baseline.json -s CRITICAL,HIGH

test-auth:
	cd ../pucora-apikeys && go test ./...
	cd ../pucora-basicauth && go test ./...
	cd ../pucora-gcp-auth && go test ./...
	cd ../pucora-aws-sigv4 && go test ./...
	cd ../pucora-ntlm && go test ./...
	cd ../pucora-revoker && go test ./...
	cd ../pucora-jwk-aggregator && go test ./...

check-fixtures-endpoints: build
	./${BIN_NAME} check -c tests/fixtures/jmespath.json
	./${BIN_NAME} check -c tests/fixtures/response-body.json
	./${BIN_NAME} check -c tests/fixtures/request-body-extractor.json
	./${BIN_NAME} check -c tests/fixtures/response-json-schema.json
	./${BIN_NAME} check -c tests/fixtures/security-policies.json
	./${BIN_NAME} check -c tests/fixtures/wildcard.json
	./${BIN_NAME} check -c tests/fixtures/ratelimit-redis.json
	./${BIN_NAME} check -c tests/fixtures/ratelimit-tiered.json
	./${BIN_NAME} check -c tests/fixtures/openapi-export.json
	./${BIN_NAME} check -c tests/fixtures/postman-export.json
	./${BIN_NAME} check -c tests/fixtures/middleware-plugin.json

test-endpoints:
	cd ../pucora-jmespath && go test ./...
	cd ../pucora-response-body && go test ./...
	cd ../pucora-request-body && go test ./...
	cd ../pucora-jsonschema && go test ./...
	cd ../pucora-security-policies && go test ./...
	cd ../pucora-wildcard && go test ./...
	cd ../pucora-ratelimit && go test ./...
	cd ../pucora-openapi && go test ./...
	cd ../pucora-postman && go test ./...
	cd ../pucora-middleware-plugin && go test ./...
	go test ./tests -run TestEndpoints -count=1 -timeout 2m

jwk-aggregator-plugin:
	@mkdir -p plugins
	cd ../pucora-jwk-aggregator/plugin && go build -buildmode=plugin -o ../../pucora-ce/plugins/jwk-aggregator.so .

check-grpc-fixtures: build
	./${BIN_NAME} check -c tests/fixtures/grpc_client.json
	./${BIN_NAME} check -c tests/fixtures/grpc_client_mapping.json
	./${BIN_NAME} check -c tests/fixtures/grpc_server.json
	./${BIN_NAME} check -c tests/fixtures/grpc_server_mixed.json
	./${BIN_NAME} check -c tests/fixtures/grpc_server_jwt.json

ws-compose-up:
	cd examples/websocket && docker compose up --build -d

ws-compose-down:
	cd examples/websocket && docker compose down

ws-compose-smoke:
	chmod +x examples/websocket/scripts/smoke.sh
	./examples/websocket/scripts/smoke.sh

ws-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off go mod vendor
	cd examples/websocket/mock-backend && GOWORK=off go mod vendor
	cd examples/websocket && docker compose up --build -d
	chmod +x examples/websocket/scripts/smoke.sh
	./examples/websocket/scripts/smoke.sh
	cd examples/websocket && docker compose down -v

graphql-compose-up:
	cd examples/graphql && docker compose up --build -d

graphql-compose-down:
	cd examples/graphql && docker compose down

graphql-compose-smoke:
	chmod +x examples/graphql/scripts/smoke.sh
	./examples/graphql/scripts/smoke.sh

graphql-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off go mod vendor
	cd examples/graphql && docker compose up --build -d
	chmod +x examples/graphql/scripts/smoke.sh
	./examples/graphql/scripts/smoke.sh
	cd examples/graphql && docker compose down -v

soap-compose-up:
	cd examples/soap && docker compose up --build -d

soap-compose-down:
	cd examples/soap && docker compose down

soap-compose-smoke:
	chmod +x examples/soap/scripts/smoke.sh
	./examples/soap/scripts/smoke.sh

soap-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/soap && docker compose up --build -d
	chmod +x examples/soap/scripts/smoke.sh
	./examples/soap/scripts/smoke.sh
	cd examples/soap && docker compose down -v

soap-compose-wssec-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/soap && PUCORA_CONFIG=pucora-wssec.json docker compose up --build -d
	chmod +x examples/soap/scripts/smoke-wssec.sh
	./examples/soap/scripts/smoke-wssec.sh
	cd examples/soap && docker compose down -v

grpc-compose-up:
	cd examples/grpc && docker compose up --build -d

grpc-compose-down:
	cd examples/grpc && docker compose down

grpc-compose-smoke:
	chmod +x examples/grpc/scripts/*.sh
	./examples/grpc/scripts/smoke-client.sh

grpc-compose-server-smoke:
	chmod +x examples/grpc/scripts/*.sh
	./examples/grpc/scripts/smoke-server.sh

grpc-compose-mixed-smoke:
	chmod +x examples/grpc/scripts/*.sh
	./examples/grpc/scripts/smoke-mixed.sh

grpc-compose-jwt-smoke:
	chmod +x examples/grpc/scripts/*.sh
	./examples/grpc/scripts/smoke-jwt.sh

grpc-compose-test: cmd/pucora-ce/schema/schema.json
	@command -v grpcurl >/dev/null 2>&1 || GOPROXY=direct go install github.com/fullstorydev/grpcurl/cmd/grpcurl@v1.9.3
	GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/grpc/mock-backend && GOWORK=off go mod tidy && GOWORK=off go mod vendor
	cd examples/grpc/mock-jwt && GOWORK=off go mod tidy && GOWORK=off go mod vendor
	cd examples/grpc && docker compose down -v 2>/dev/null || true
	cd examples/grpc && docker compose up --build -d
	chmod +x examples/grpc/scripts/*.sh
	./examples/grpc/scripts/smoke-client.sh
	cd examples/grpc && docker compose down -v
	cd examples/grpc && PUCORA_CONFIG=pucora-server.json docker compose up --build -d
	./examples/grpc/scripts/smoke-server.sh
	cd examples/grpc && docker compose down -v
	cd examples/grpc && PUCORA_CONFIG=pucora-mixed.json docker compose up --build -d
	./examples/grpc/scripts/smoke-mixed.sh
	cd examples/grpc && docker compose down -v
	cd examples/grpc && PUCORA_CONFIG=pucora-jwt.json docker compose up --build -d
	./examples/grpc/scripts/smoke-jwt.sh
	cd examples/grpc && docker compose down -v

pubsub-nats-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/nats && docker compose up --build -d
	chmod +x examples/pubsub/nats/scripts/smoke.sh examples/pubsub/scripts/common.sh
	./examples/pubsub/nats/scripts/smoke.sh
	cd examples/pubsub/nats && docker compose down -v

pubsub-kafka-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/kafka && docker compose up --build -d
	chmod +x examples/pubsub/kafka/scripts/smoke.sh examples/pubsub/scripts/common.sh
	./examples/pubsub/kafka/scripts/smoke.sh
	cd examples/pubsub/kafka && docker compose down -v

pubsub-rabbit-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/rabbit && docker compose up --build -d
	chmod +x examples/pubsub/rabbit/scripts/smoke.sh examples/pubsub/scripts/common.sh
	./examples/pubsub/rabbit/scripts/smoke.sh
	cd examples/pubsub/rabbit && docker compose down -v

pubsub-gcp-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/gcp && docker compose up --build -d || (docker compose logs pubsub-init; exit 1)
	chmod +x examples/pubsub/gcp/scripts/smoke.sh examples/pubsub/gcp/scripts/init-pubsub.sh examples/pubsub/scripts/common.sh
	./examples/pubsub/gcp/scripts/smoke.sh
	cd examples/pubsub/gcp && docker compose down -v

pubsub-aws-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/aws && docker compose up --build -d
	chmod +x examples/pubsub/aws/scripts/smoke.sh examples/pubsub/scripts/common.sh
	./examples/pubsub/aws/scripts/smoke.sh
	cd examples/pubsub/aws && docker compose down -v

pubsub-azure-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/azure && docker compose up --build -d
	chmod +x examples/pubsub/azure/scripts/smoke.sh examples/pubsub/scripts/common.sh
	./examples/pubsub/azure/scripts/smoke.sh
	cd examples/pubsub/azure && docker compose down -v

pubsub-kafka-advanced-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/kafka-advanced && docker compose up --build -d
	chmod +x examples/pubsub/kafka-advanced/scripts/smoke.sh examples/pubsub/scripts/common.sh
	./examples/pubsub/kafka-advanced/scripts/smoke.sh
	cd examples/pubsub/kafka-advanced && docker compose down -v

pubsub-async-kafka-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/pucora/* GOSUMDB=off go mod vendor
	cd examples/pubsub/async-kafka && docker compose up --build -d
	chmod +x examples/pubsub/async-kafka/scripts/smoke.sh
	./examples/pubsub/async-kafka/scripts/smoke.sh
	cd examples/pubsub/async-kafka && docker compose down -v

sse-compose-up:
	cd examples/streaming && docker compose up --build -d

sse-compose-down:
	cd examples/streaming && docker compose down

sse-compose-smoke:
	chmod +x examples/streaming/scripts/smoke.sh
	./examples/streaming/scripts/smoke.sh

sse-compose-test: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off go mod vendor
	cd examples/streaming/mock-backend && GOWORK=off go mod vendor
	cd examples/streaming && docker compose up --build -d
	chmod +x examples/streaming/scripts/smoke.sh
	./examples/streaming/scripts/smoke.sh
	cd examples/streaming && docker compose down -v

SCHEMA_URL := https://raw.githubusercontent.com/pucora/pucora-schema/v2.0.4/v2.13/pucora.json

cmd/pucora-ce/schema/schema.json:
	@echo "Fetching v${SCHEMA_VERSION} schema"
	@mkdir -p $(dir $@)
	@cp ../pucora-schema/v${SCHEMA_VERSION}/pucora.json $@ 2>/dev/null || \
		curl -fsSL -o $@ $(SCHEMA_URL)

# Build Pucora using docker (defaults to whatever the golang container uses)
build_on_docker: docker-builder-linux
	docker run --rm -it -v "${PWD}:/app" -w /app $(DOCKER_BUILDER):${VERSION}-linux-generic sh -c "git config --global --add safe.directory /app && make -e build"

# Build the container using the Dockerfile (alpine)
docker: cmd/pucora-ce/schema/schema.json
	@test -d vendor || GOWORK=off go mod vendor
	docker build --pull \
		--build-arg GOLANG_VERSION=${GOLANG_VERSION} \
		--build-arg ALPINE_VERSION=${ALPINE_VERSION} \
		--build-arg VERSION=${VERSION} \
		-t $(DOCKER_CE):${VERSION} .

CHART_DIR := deploy/helm/pucora

.PHONY: sync-chart-version verify-chart-version helm-lint helm-cluster-test

sync-chart-version:
	@deploy/helm/scripts/sync-chart-version.sh $(VERSION)

verify-chart-version:
	@deploy/helm/scripts/verify-chart-version.sh $(VERSION)

helm-lint:
	helm lint $(CHART_DIR)
	helm lint $(CHART_DIR) -f $(CHART_DIR)/ci/values-prod.yaml
	helm lint $(CHART_DIR) -f $(CHART_DIR)/ci/values-aws-nlb.yaml
	helm lint $(CHART_DIR) -f $(CHART_DIR)/ci/values-istio.yaml
	@deploy/helm/scripts/verify-chart-version.sh $(VERSION)

helm-cluster-test:
	@deploy/helm/scripts/cluster-test.sh

docker-builder:
	docker build --no-cache --pull --build-arg GOLANG_VERSION=${GOLANG_VERSION} --build-arg ALPINE_VERSION=${ALPINE_VERSION} -t $(DOCKER_BUILDER):${VERSION} -f Dockerfile-builder .

docker-builder-linux:
	docker build --no-cache --pull --build-arg GOLANG_VERSION=${GOLANG_VERSION} -t $(DOCKER_BUILDER):${VERSION}-linux-generic -f Dockerfile-builder-linux .

benchmark:
	@mkdir -p bench_res
	@touch bench_res/${GIT_COMMIT}.out
	@docker run --rm -d --name pucora -v "${PWD}/tests/fixtures:/etc/pucora" -p 8080:8080 $(DOCKER_CE):${VERSION} run -dc /etc/pucora/bench.json
	@sleep 2
	@docker run --rm -it --link pucora peterevans/vegeta sh -c \
		"echo 'GET http://pucora:8080/test' | vegeta attack -rate=0 -duration=30s -max-workers=300 | tee results.bin | vegeta report" > bench_res/${GIT_COMMIT}.out
	@docker stop pucora
	@cat bench_res/${GIT_COMMIT}.out

security_scan:
	@mkdir -p sec_scan
	@touch sec_scan/${GIT_COMMIT}.out
	@docker run --rm -d --name pucora -v "${PWD}/tests/fixtures:/etc/pucora" -p 8080:8080 $(DOCKER_CE):${VERSION} run -dc /etc/pucora/bench.json
	@docker run --rm -it --link pucora instrumentisto/nmap --script vuln pucora > sec_scan/${GIT_COMMIT}.out
	@docker stop pucora
	@cat sec_scan/${GIT_COMMIT}.out

builder/skel/%/etc/init.d/pucora: builder/files/pucora.init
	mkdir -p "$(dir $@)"
	cp builder/files/pucora.init "$@"

builder/skel/%/usr/bin/pucora: pucora
	mkdir -p "$(dir $@)"
	cp pucora "$@"

builder/skel/%/etc/pucora/pucora.json: pucora.json
	mkdir -p "$(dir $@)"
	cp pucora.json "$@"

builder/skel/%/lib/systemd/system/pucora.service: builder/files/pucora.service
	mkdir -p "$(dir $@)"
	cp builder/files/pucora.service "$@"

builder/skel/%/usr/lib/systemd/system/pucora.service: builder/files/pucora.service
	mkdir -p "$(dir $@)"
	cp builder/files/pucora.service "$@"

builder/skel/%/etc/rsyslog.d/pucora.conf: builder/files/pucora.conf-rsyslog
	mkdir -p "$(dir $@)"
	cp builder/files/pucora.conf-rsyslog "$@"

builder/skel/%/etc/logrotate.d/pucora: builder/files/pucora-logrotate
	mkdir -p "$(dir $@)"
	cp builder/files/pucora-logrotate "$@"

.PHONY: tgz
tgz: builder/skel/tgz/usr/bin/pucora
tgz: builder/skel/tgz/etc/pucora/pucora.json
tgz: builder/skel/tgz/etc/init.d/pucora
	tar zcvf pucora_${VERSION}_${ARCH}${OS_TAG}.tar.gz -C builder/skel/tgz/ .

.PHONY: deb
deb: builder/skel/deb/usr/bin/pucora
deb: builder/skel/deb/etc/pucora/pucora.json
deb: builder/skel/deb/etc/rsyslog.d/pucora.conf
deb: builder/skel/deb/etc/logrotate.d/pucora
	docker run --rm -it -v "${PWD}:${DOCKER_WDIR}" -w ${DOCKER_WDIR} ${DOCKER_FPM}:deb -t deb ${DEB_OPTS} \
		--iteration ${RELEASE} \
		--deb-systemd builder/files/pucora.service \
		-C builder/skel/deb \
		${FPM_OPTS}

.PHONY: rpm
rpm: builder/skel/rpm/usr/lib/systemd/system/pucora.service
rpm: builder/skel/rpm/usr/bin/pucora
rpm: builder/skel/rpm/etc/pucora/pucora.json
rpm: builder/skel/rpm/etc/rsyslog.d/pucora.conf
rpm: builder/skel/rpm/etc/logrotate.d/pucora
	docker run --rm -it -v "${PWD}:${DOCKER_WDIR}" -w ${DOCKER_WDIR} ${DOCKER_FPM}:rpm -t rpm ${RPM_OPTS} \
		--iteration ${RELEASE} \
		-C builder/skel/rpm \
		${FPM_OPTS}

.PHONY: deb-release
deb-release: builder/skel/deb-release/usr/bin/pucora
deb-release: builder/skel/deb-release/etc/pucora/pucora.json
	/usr/local/bin/fpm -t deb ${DEB_OPTS} \
		--iteration ${RELEASE} \
		--deb-systemd builder/files/pucora.service \
		-C builder/skel/deb-release \
		${FPM_OPTS}

.PHONY: rpm-release
rpm-release: builder/skel/rpm-release/usr/lib/systemd/system/pucora.service
rpm-release: builder/skel/rpm-release/usr/bin/pucora
rpm-release: builder/skel/rpm-release/etc/pucora/pucora.json
	/usr/local/bin/fpm -t rpm ${RPM_OPTS} \
		--iteration ${RELEASE} \
		-C builder/skel/rpm-release \
		${FPM_OPTS}

.PHONY: clean
clean:
	rm -rf builder/skel/*
	rm -f ${BIN_NAME}
	rm -rf vendor/
	rm -f cmd/pucora-ce/schema/schema.json
