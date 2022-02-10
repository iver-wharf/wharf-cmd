.PHONY: install tidy deps \
	lint lint-md lint-go \
	lint-fix lint-md-fix

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
	npm install

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
