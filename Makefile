.PHONY: all build test

# This Makefile is a simple example that demonstrates usual steps to build a binary that can be run in the same
# architecture that was compiled in. The "ldflags" in the build assure that any needed dependency is included in the
# binary and no external dependencies are needed to run the service.

BIN_NAME :=velonetics
OS := $(shell uname | tr '[:upper:]' '[:lower:]')
MODULE := github.com/velonetics/velonetics-ce/v2
VERSION := 2.0.1
SCHEMA_VERSION := 2.13
GIT_COMMIT := $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo "unknown")
PKGNAME := velonetics
LICENSE := Apache 2.0
VENDOR=
URL := https://velonetics.io
RELEASE := 0
USER := velonetics
ARCH := amd64
DESC := High performance API gateway. Aggregate, filter, manipulate and add middlewares
MAINTAINER := Velonetics Team <community@velonetics.io>
DOCKER_WDIR := /tmp/fpm
DOCKER_FPM := velonetics/fpm
GOLANG_VERSION := 1.25.11
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

all: test test-websocket test-soap

build: cmd/velonetics-ce/schema/schema.json
	@echo "Building the binary..."
	@go get .
	@go build -ldflags="-X ${MODULE}/pkg.Version=${VERSION} -X github.com/velonetics/lura/v2/core.VeloneticsVersion=${VERSION} \
	-X github.com/velonetics/lura/v2/core.GlibcVersion=${GLIBC_VERSION} ${EXTRA_LDFLAGS}" \
	-o ${BIN_NAME} ./cmd/velonetics-ce
	@echo "You can now use ./${BIN_NAME}"

test: build
	go test -v ./tests

test-websocket:
	cd ../velonetics-websocket && go test ./...

test-soap:
	cd ../velonetics-soap && go test ./...

check-fixtures: build
	./${BIN_NAME} check -c tests/fixtures/ws_direct.json
	./${BIN_NAME} check -c tests/fixtures/ws_multiplex.json
	./${BIN_NAME} check -c tests/fixtures/ws_jwt.json
	./${BIN_NAME} check -c velonetics-ws.json
	./${BIN_NAME} check -c tests/fixtures/soap_country.json

ws-compose-up:
	cd examples/websocket && docker compose up --build -d

ws-compose-down:
	cd examples/websocket && docker compose down

ws-compose-smoke:
	chmod +x examples/websocket/scripts/smoke.sh
	./examples/websocket/scripts/smoke.sh

ws-compose-test: cmd/velonetics-ce/schema/schema.json
	@test -d vendor || GOWORK=off go mod vendor
	cd examples/websocket/mock-backend && GOWORK=off go mod vendor
	cd examples/websocket && docker compose up --build -d
	chmod +x examples/websocket/scripts/smoke.sh
	./examples/websocket/scripts/smoke.sh
	cd examples/websocket && docker compose down -v

soap-compose-up:
	cd examples/soap && docker compose up --build -d

soap-compose-down:
	cd examples/soap && docker compose down

soap-compose-smoke:
	chmod +x examples/soap/scripts/smoke.sh
	./examples/soap/scripts/smoke.sh

soap-compose-test: cmd/velonetics-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/velonetics/* GOSUMDB=off go mod vendor
	cd examples/soap && docker compose up --build -d
	chmod +x examples/soap/scripts/smoke.sh
	./examples/soap/scripts/smoke.sh
	cd examples/soap && docker compose down -v

soap-compose-wssec-test: cmd/velonetics-ce/schema/schema.json
	@test -d vendor || GOWORK=off GOPROXY=direct GOPRIVATE=github.com/velonetics/* GOSUMDB=off go mod vendor
	cd examples/soap && VELO_CONFIG=velonetics-wssec.json docker compose up --build -d
	chmod +x examples/soap/scripts/smoke-wssec.sh
	./examples/soap/scripts/smoke-wssec.sh
	cd examples/soap && docker compose down -v

SCHEMA_URL := https://raw.githubusercontent.com/velonetics/velonetics-schema/v2.0.2/v2.13/velonetics.json

cmd/velonetics-ce/schema/schema.json:
	@echo "Fetching v${SCHEMA_VERSION} schema"
	@mkdir -p $(dir $@)
	@cp ../velonetics-schema/v${SCHEMA_VERSION}/velonetics.json $@ 2>/dev/null || \
		curl -fsSL -o $@ $(SCHEMA_URL)

# Build Velonetics using docker (defaults to whatever the golang container uses)
build_on_docker: docker-builder-linux
	docker run --rm -it -v "${PWD}:/app" -w /app velonetics/builder:${VERSION}-linux-generic sh -c "git config --global --add safe.directory /app && make -e build"

# Build the container using the Dockerfile (alpine)
docker: cmd/velonetics-ce/schema/schema.json
	@test -d vendor || GOWORK=off go mod vendor
	docker build --pull \
		--build-arg GOLANG_VERSION=${GOLANG_VERSION} \
		--build-arg ALPINE_VERSION=${ALPINE_VERSION} \
		--build-arg VERSION=${VERSION} \
		-t niteesh20/velonetics:${VERSION} .

docker-builder:
	docker build --no-cache --pull --build-arg GOLANG_VERSION=${GOLANG_VERSION} --build-arg ALPINE_VERSION=${ALPINE_VERSION} -t velonetics/builder:${VERSION} -f Dockerfile-builder .

docker-builder-linux:
	docker build --no-cache --pull --build-arg GOLANG_VERSION=${GOLANG_VERSION} -t velonetics/builder:${VERSION}-linux-generic -f Dockerfile-builder-linux .

benchmark:
	@mkdir -p bench_res
	@touch bench_res/${GIT_COMMIT}.out
	@docker run --rm -d --name velonetics -v "${PWD}/tests/fixtures:/etc/velonetics" -p 8080:8080 velonetics/velonetics:${VERSION} run -dc /etc/velonetics/bench.json
	@sleep 2
	@docker run --rm -it --link velonetics peterevans/vegeta sh -c \
		"echo 'GET http://velonetics:8080/test' | vegeta attack -rate=0 -duration=30s -max-workers=300 | tee results.bin | vegeta report" > bench_res/${GIT_COMMIT}.out
	@docker stop velonetics
	@cat bench_res/${GIT_COMMIT}.out

security_scan:
	@mkdir -p sec_scan
	@touch sec_scan/${GIT_COMMIT}.out
	@docker run --rm -d --name velonetics -v "${PWD}/tests/fixtures:/etc/velonetics" -p 8080:8080 velonetics/velonetics:${VERSION} run -dc /etc/velonetics/bench.json
	@docker run --rm -it --link velonetics instrumentisto/nmap --script vuln velonetics > sec_scan/${GIT_COMMIT}.out
	@docker stop velonetics
	@cat sec_scan/${GIT_COMMIT}.out

builder/skel/%/etc/init.d/velonetics: builder/files/velonetics.init
	mkdir -p "$(dir $@)"
	cp builder/files/velonetics.init "$@"

builder/skel/%/usr/bin/velonetics: velonetics
	mkdir -p "$(dir $@)"
	cp velonetics "$@"

builder/skel/%/etc/velonetics/velonetics.json: velonetics.json
	mkdir -p "$(dir $@)"
	cp velonetics.json "$@"

builder/skel/%/lib/systemd/system/velonetics.service: builder/files/velonetics.service
	mkdir -p "$(dir $@)"
	cp builder/files/velonetics.service "$@"

builder/skel/%/usr/lib/systemd/system/velonetics.service: builder/files/velonetics.service
	mkdir -p "$(dir $@)"
	cp builder/files/velonetics.service "$@"

builder/skel/%/etc/rsyslog.d/velonetics.conf: builder/files/velonetics.conf-rsyslog
	mkdir -p "$(dir $@)"
	cp builder/files/velonetics.conf-rsyslog "$@"

builder/skel/%/etc/logrotate.d/velonetics: builder/files/velonetics-logrotate
	mkdir -p "$(dir $@)"
	cp builder/files/velonetics-logrotate "$@"

.PHONY: tgz
tgz: builder/skel/tgz/usr/bin/velonetics
tgz: builder/skel/tgz/etc/velonetics/velonetics.json
tgz: builder/skel/tgz/etc/init.d/velonetics
	tar zcvf velonetics_${VERSION}_${ARCH}${OS_TAG}.tar.gz -C builder/skel/tgz/ .

.PHONY: deb
deb: builder/skel/deb/usr/bin/velonetics
deb: builder/skel/deb/etc/velonetics/velonetics.json
deb: builder/skel/deb/etc/rsyslog.d/velonetics.conf
deb: builder/skel/deb/etc/logrotate.d/velonetics
	docker run --rm -it -v "${PWD}:${DOCKER_WDIR}" -w ${DOCKER_WDIR} ${DOCKER_FPM}:deb -t deb ${DEB_OPTS} \
		--iteration ${RELEASE} \
		--deb-systemd builder/files/velonetics.service \
		-C builder/skel/deb \
		${FPM_OPTS}

.PHONY: rpm
rpm: builder/skel/rpm/usr/lib/systemd/system/velonetics.service
rpm: builder/skel/rpm/usr/bin/velonetics
rpm: builder/skel/rpm/etc/velonetics/velonetics.json
rpm: builder/skel/rpm/etc/rsyslog.d/velonetics.conf
rpm: builder/skel/rpm/etc/logrotate.d/velonetics
	docker run --rm -it -v "${PWD}:${DOCKER_WDIR}" -w ${DOCKER_WDIR} ${DOCKER_FPM}:rpm -t rpm ${RPM_OPTS} \
		--iteration ${RELEASE} \
		-C builder/skel/rpm \
		${FPM_OPTS}

.PHONY: deb-release
deb-release: builder/skel/deb-release/usr/bin/velonetics
deb-release: builder/skel/deb-release/etc/velonetics/velonetics.json
	/usr/local/bin/fpm -t deb ${DEB_OPTS} \
		--iteration ${RELEASE} \
		--deb-systemd builder/files/velonetics.service \
		-C builder/skel/deb-release \
		${FPM_OPTS}

.PHONY: rpm-release
rpm-release: builder/skel/rpm-release/usr/lib/systemd/system/velonetics.service
rpm-release: builder/skel/rpm-release/usr/bin/velonetics
rpm-release: builder/skel/rpm-release/etc/velonetics/velonetics.json
	/usr/local/bin/fpm -t rpm ${RPM_OPTS} \
		--iteration ${RELEASE} \
		-C builder/skel/rpm-release \
		${FPM_OPTS}

.PHONY: clean
clean:
	rm -rf builder/skel/*
	rm -f ${BIN_NAME}
	rm -rf vendor/
	rm -f cmd/velonetics-ce/schema/schema.json
