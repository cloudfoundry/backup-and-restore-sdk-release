###### Help ###################################################################

.DEFAULT_GOAL = help

.PHONY: help

help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Targets ################################################################

test: fmt vet ginkgo ## Runs all build, static analysis, and test steps

vet: ## Run static code analysis
	go vet ./...

ginkgo: ## Run tests
	go run github.com/onsi/ginkgo/v2/ginkgo -r

fmt: ## Checks that the code is formatted correctly
	@@if [ -n "$$(gofmt -s -e -l -d .)" ]; then \
		echo "gofmt check failed: run 'gofmt -s -e -l -w .'"; \
		exit 1; \
	fi