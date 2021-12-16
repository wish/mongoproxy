BUILD := build
GO ?= go
GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*")
GOFMT ?= gofmt
GOIMPORTS ?= goimports -local=github.com/wish/mongoproxy
STATICCHECK ?= staticcheck
VERSION := $(shell git describe --tags 2> /dev/null || echo "unreleased")
V_DIRTY := $(shell git describe --tags --exact-match HEAD 2> /dev/null > /dev/null || echo "-unreleased")
GIT     := $(shell git rev-parse --short HEAD)
DIRTY   := $(shell git diff-index --quiet HEAD 2> /dev/null > /dev/null || echo "-dirty")

.PHONY: clean
clean:
	$(GO) clean -i ./...
	rm -rf $(BUILD)

.PHONY: static-check
static-check:
	$(STATICCHECK) ./...

.PHONY: fmt
fmt:
	$(GOFMT) -w -s $(GOFILES)

.PHONY: imports
imports:
	$(GOIMPORTS) -w $(GOFILES)

.PHONY: test
test:
	cd pkg && $(GO) test -mod=vendor ./...

.PHONY: integrationtest
integrationtest:
	cd integrationtest && $(GO) test -mod=vendor -v

.PHONY: vendor
vendor:
	GO111MODULE=on $(GO) mod tidy
	GO111MODULE=on $(GO) mod vendor

# TODO: cleanup
go-client-tests:
	cd /tmp && git clone git@github.com:mongodb/mongo-go-driver || exit 0
	cd /tmp/mongo-go-driver/mongo/integration && MONGODB_URI="mongodb://localhost:27016" go test .

python-client-tests:
	cd /tmp && git clone https://github.com/mongodb/mongo-python-driver.git
	cd /tmp/mongo-python-driver && DB_PORT=27016 python3 setup.py test --xunit-output test.output

OLD-python-client-tests:
	cd /tmp && git clone https://github.com/mongodb/mongo-python-driver.git && git checkout 2.8
	cd /tmp/mongo-python-driver && DB_PORT=27016 python2 setup.py

.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 docker build .

testlocal-build:
	DOCKER_BUILDKIT=1 docker build -t quay.io/wish/mongoproxy:latest .
	kind load docker-image quay.io/wish/mongoproxy:latest
