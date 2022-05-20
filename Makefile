
commit = $(shell git rev-parse HEAD)
version = latest

ifeq ($(OS),Windows_NT)
wharf.exe: swag
	go build -o wharf.exe
else
wharf: swag
	go build -o wharf
endif

.PHONY: clean
clean: clean-swag clean-build

.PHONY: clean-build
clean-build:
ifeq ($(OS),Windows_NT)
	rm -rfv wharf.exe
else
	rm -rfv wharf
endif

.PHONY: install
install: swag
	go install ./cmd/wharf

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: deps
deps: deps-go deps-npm

.PHONY: deps-go
deps-go:
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	go install github.com/alta/protopatch/cmd/protoc-gen-go-patch@v0.5.0
	go install github.com/swaggo/swag/cmd/swag@v1.8.1
	go install github.com/yoheimuta/protolint/cmd/protolint@v0.37.1

.PHONY: deps-npm
deps-npm:
	npm install

.PHONY: check
check: swag
	go test ./...

.PHONY: proto
proto: api/workerapi/v1/worker.pb.go

api/workerapi/v1/worker.pb.go: api/workerapi/v1/worker.proto
	protoc -I . \
		-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
		-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
		--go-patch_out=plugin=go,paths=source_relative:. \
		--go-patch_out=plugin=go-grpc,paths=source_relative:. \
		./api/workerapi/v1/worker.proto
# Generated files have some non-standard formatting, so let's format it.
	goimports -w ./api/workerapi/v1/.

.PHONY: docker
docker:
	docker build . \
		--pull \
		-t "quay.io/iver-wharf/wharf-cmd:latest" \
		-t "quay.io/iver-wharf/wharf-cmd:$(version)" \
		--build-arg BUILD_VERSION="$(version)" \
		--build-arg BUILD_GIT_COMMIT="$(commit)" \
		--build-arg BUILD_DATE="$(shell date --iso-8601=seconds)"
	@echo ""
	@echo "Push the image by running:"
	@echo "docker push quay.io/iver-wharf/wharf-cmd:latest"
ifneq "$(version)" "latest"
	@echo "docker push quay.io/iver-wharf/wharf-cmd:$(version)"
endif

.PHONY: docker-run
docker-run:
	docker run --rm -it quay.io/iver-wharf/wharf-cmd:$(version)

.PHONY: clean-swag
clean-swag:
	rm -vrf pkg/provisioner/provisionerserver/docs pkg/workerapi/workerserver/docs

.PHONY: swag-force
swag-force: clean-swag swag

.PHONY: swag
swag: \
	pkg/provisioner/provisionerserver/docs \
	pkg/workerapi/workerserver/docs

pkg/provisioner/provisionerserver/docs:
	swag init \
		--dir pkg/provisioner/provisionerserver,pkg/provisioner,pkg/worker \
		--generalInfo provisionerapi.go \
		--output pkg/provisioner/provisionerserver/docs \
		--instanceName provisionerapi

pkg/workerapi/workerserver/docs:
	swag init \
		--dir pkg/workerapi/workerserver \
		--parseDependency --parseDepth 2 \
		--generalInfo restserver.go \
		--output pkg/workerapi/workerserver/docs \
		--instanceName workerapi

.PHONY: lint lint-fix \
	lint-md lint-go lint-proto \
	lint-fix-md lint-fix-go lint-fix-proto
lint: lint-md lint-go lint-proto
lint-fix: lint-fix-md lint-fix-go lint-fix-proto

lint-md:
	npx remark . .github

lint-fix-md:
	npx remark . .github -o

lint-go:
	@echo goimports -d '**/*.go'
	@goimports -d $(shell git ls-files "*.go")
	revive -formatter stylish -config revive.toml ./...

lint-fix-go:
	@echo goimports -d -w '**/*.go'
	@goimports -d -w $(shell git ls-files "*.go")

lint-proto:
	protolint lint api/workerapi

lint-fix-proto:
	protolint lint -fix api/workerapi
