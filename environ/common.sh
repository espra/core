# Public Domain (-) 2016-present, The Core Authors.
# See the Core UNLICENSE file for details.

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

trap log_error ERR

function exists() {
    command -v $1 >/dev/null 2>&1
}

function log_error() {
    if [[ "$#" -eq 0 ]]; then
        printf "\n\033[31;1m!! ERROR:\n!! ERROR: failed to run ${0}\n!! ERROR:\033[0m\n"
    else
        printf "\n\033[31;1m!! ERROR:\n!! ERROR: failed to run ${0}: ${@}\n!! ERROR:\033[0m\n"
    fi
}

function log_fatal() {
    log_error "$@"
    exit 1
}

function print_progress() {
    printf "  \033[32;1m✨ \033[0m ${@}\n"
}

function print_success() {
    printf "  \033[32;1m✔  \033[0m ${@}\n"
}
