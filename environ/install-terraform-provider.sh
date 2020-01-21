#! /usr/bin/env bash

# Public Domain (-) 2020-present, The Core Authors.
# See the Core UNLICENSE file for details.

ENVIRON_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
source "${ENVIRON_DIR}/common.sh"

cd "${ENVIRON_DIR}/../infra/terraform"

print_progress "Building terraform-provider-core"
go build -o provider.bin ./provider
OUTFILE=terraform-provider-core.v$(./provider.bin version)

mkdir -p ~/.terraform.d/plugins
mv provider.bin ~/.terraform.d/plugins/${OUTFILE}
print_success "Successfully installed ${OUTFILE}"
