# Public Domain (-) 2020-present, The Core Authors.
# See the Core UNLICENSE file for details.

.DEFAULT_GOAL := help

.PHONY: help install-plugins

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-plugins: ## install terraform plugins
	@mkdir -p ~/.terraform.d/plugins; \
	go build -o provider.bin ./terraform-provider-core; \
	VERSION=$$(./provider.bin version); \
	mv provider.bin ~/.terraform.d/plugins/terraform-provider-core.v$${VERSION}; \
	echo ">> Successfully installed: terraform-provider-core.v$${VERSION}"
