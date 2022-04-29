.PHONY: install tidy deps deps-go deps-npm check \
	docker docker-run swag swag-force proto \
	lint lint-md lint-go \
	lint-fix lint-fix-md lint-fix-go

commit = $(shell git rev-parse HEAD)
version = latest

ifeq ($(OS),Windows_NT)
wharf.exe: swag
	go build -o wharf.exe
else
wharf: swag
	go build -o wharf
endif

install: swag
	go install ./cmd/wharf

tidy:
	go mod tidy

deps: deps-go deps-npm

deps-go:
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	go install github.com/alta/protopatch/cmd/protoc-gen-go-patch@v0.5.0
	go install github.com/swaggo/swag/cmd/swag@v1.8.0

deps-npm:
	npm install

check:
	go test ./...

proto:
	protoc -I . \
		-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
		-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
		--go-patch_out=plugin=go,paths=source_relative:. \
		--go-patch_out=plugin=go-grpc,paths=source_relative:. \
		./api/workerapi/v1/worker.proto
# Generated files have some non-standard formatting, so let's format it.
	goimports -w ./api/workerapi/v1/.

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

docker-run:
	docker run --rm -it quay.io/iver-wharf/wharf-api:$(version)

swag-force:
	swag init \
		--dir pkg/provisionerapi,pkg/provisioner,pkg/worker \
		--generalInfo provisionerapi.go --output pkg/provisionerapi/docs

swag: pkg/provisionerapi/docs/docs.go

pkg/provisionerapi/docs/docs.go:
	swag init \
		--dir pkg/provisionerapi,pkg/provisioner,pkg/worker \
		--generalInfo provisionerapi.go --output pkg/provisionerapi/docs

lint: lint-md lint-go
lint-fix: lint-fix-md lint-fix-go

lint-md:
	npx remark . .github

lint-fix-md:
	npx remark . .github -o

lint-go:
	goimports -d $(shell git ls-files "*.go")
	revive -formatter stylish -config revive.toml ./...

lint-fix-go:
	goimports -d -w $(shell git ls-files "*.go")
