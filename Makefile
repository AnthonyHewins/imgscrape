.DEFAULT: cli
.PHONY: fmt test gen clean run help

# command aliases
test := CONFIG_ENV=test go test ./...

VERSION ?= v0.0.0
build_flag_path := github.com/AnthonyHewins/imgscrape
BUILD_FLAGS := 
ifneq (,$(wildcard ./vendor))
	$(info Found vendor directory; setting "-mod vendor" to any "go build" commands)
	BUILD_FLAGS += -mod vendor
endif

cli:
	go build $(BUILD_FLAGS) -ldflags="-X '$(build_flag_path)/cli/cmd.version=$(VERSION)'" -o bin/imgscrape cmd/$@/*.go

test: ## Run go vet, then test all files
	go vet ./...
	$(test)

update-snapshots: ## Update snapshots during a go test. Must have cupaloy
	UPDATE_SNAPSHOTS=true $(test)

clean: ## gofmt, go generate, then go mod tidy, and finally rm -rf bin/
	find . -iname *.go -type f -exec gofmt -w -s {} \;
	go generate ./...
	go mod tidy
	rm -rf ./bin

help: ## Print help
	@printf "\033[36m%-30s\033[0m %s\n" "(target)" "Build a target binary in current arch for running locally: $(targets)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
