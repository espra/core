# Public Domain (-) 2020-present, The Core Authors.
# See the Core UNLICENSE file for details.

.DEFAULT_GOAL := help

.PHONY: help install-terraform-provider

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-terraform-provider: ## install custom terraform provider
	@../environ/install-terraform-provider.sh
