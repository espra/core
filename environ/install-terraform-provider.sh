#! /usr/bin/env bash

# Public Domain (-) 2020-present, The Core Authors.
# See the Core UNLICENSE file for details.

ENVIRON_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
source "${ENVIRON_DIR}/common.sh"

cd "${ENVIRON_DIR}/../infra/provider"

print_progress "Building terraform-provider-core"
go build -o provider.bin ./

OSARCH=$(./provider.bin osarch)
OUTFILE=terraform-provider-core_v$(./provider.bin version)

mkdir -p "${HOME}/.terraform.d/plugins/${OSARCH}"
mv provider.bin "${HOME}/.terraform.d/plugins/${OSARCH}/${OUTFILE}"
print_success "Successfully installed ${OUTFILE}"
