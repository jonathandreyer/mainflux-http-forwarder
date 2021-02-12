# Copyright (c) J.Dreyer
# SPDX-License-Identifier: Apache-2.0

BUILD_DIR = build
CGO_ENABLED ?= 0
GOARCH ?= amd64

define compile_http_forwarder
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -mod=vendor -ldflags "-s -w" -o ${BUILD_DIR}/mainflux-http-forwarder cmd/http-forwarder/main.go
endef

define make_docker
	$(eval svc=$(subst docker_,,$(1)))

	docker build \
		--no-cache \
		--build-arg SVC=$(svc) \
		--build-arg GOARCH=$(GOARCH) \
		--build-arg GOARM=$(GOARM) \
		--tag=jonathandreyer/mainflux-$(svc) \
		-f docker/Dockerfile .
endef

define make_docker_dev
	$(eval svc=$(subst docker_dev_,,$(1)))

	docker build \
		--no-cache \
		--build-arg SVC=$(svc) \
		--tag=jonathandreyer/mainflux-$(svc) \
		-f docker/Dockerfile.dev ./build
endef

all:
	$(call compile_http_forwarder)

.PHONY: all docker docker_dev latest release

clean:
	rm -rf ${BUILD_DIR}

cleandocker:
	# Remove mainflux containers
	docker ps -f name=mainflux-http-forwarder -aq | xargs -r docker rm

	# Remove exited containers
	docker ps -f name=mainflux-http-forwarder -f status=dead -f status=exited -aq | xargs -r docker rm -v

	# Remove unused images
	docker images "mainflux-http-forwarder\/*" -f dangling=true -q | xargs -r docker rmi

	# Remove old mainflux images
	docker images -q mainflux-http-forwarder\/* | xargs -r docker rmi

ifdef pv
	# Remove unused volumes
	docker volume ls -f name=mainflux-http-forwarder -f dangling=true -q | xargs -r docker volume rm
endif

install:
	cp ${BUILD_DIR}/* $(GOBIN)

test:
	go test -mod=vendor -v -race -count 1 -tags test $(shell go list ./... | grep -v 'vendor\|cmd')

docker:
	$(call make_docker,docker_http-forwarder,$(GOARCH))
docker_dev:
	$(call make_docker_dev,docker_dev_http-forwarder)

define docker_push
	docker push jonathandreyer/mainflux-http-forwarder:$(1)
endef

changelog:
	git log $(shell git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s"

latest: docker
	$(call docker_push,latest)

release:
	$(eval version = $(shell git describe --abbrev=0 --tags))
	git checkout $(version)
	$(MAKE) docker
	docker tag jonathandreyer/mainflux-http-forwarder jonathandreyer/mainflux-http-forwarder:$(version)
	$(call docker_push,$(version))
