.PHONY: install tidy deps \
	proto lint lint-md lint-go \
	lint-fix lint-fix-md lint-fix-go

ifeq ($(OS),Windows_NT)
wharf.exe:
	go build -o wharf.exe
else
wharf:
	go build -o wharf
endif

install:
	go install

tidy:
	go mod tidy

deps:
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	go install github.com/alta/protopatch/cmd/protoc-gen-go-patch@v0.5.0
	npm install

proto:
	protoc -I . \
		-I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
		-I `go list -m -f {{.Dir}} google.golang.org/protobuf` \
		--go-patch_out=plugin=go,paths=source_relative:. \
		--go-patch_out=plugin=go-grpc,paths=source_relative:. \
		./api/workerapi/v1/worker.proto
# Generated files have some non-standard formatting, so let's format it.
	goimports -w ./api/workerapi/v1/.

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
