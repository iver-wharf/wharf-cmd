.PHONY: install tidy deps \
	swag swag-force \
	lint lint-md lint-go \
	lint-fix lint-fix-md lint-fix-go

ifeq ($(OS),Windows_NT)
wharf.exe: swag
	go build -o wharf.exe
else
wharf: swag
	go build -o wharf
endif

install: swag
	go install

tidy:
	go mod tidy

deps:
	go install github.com/mgechev/revive@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@v1.8.0
	npm install

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
