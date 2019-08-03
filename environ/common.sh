# Public Domain (-) 2018-present, The Amp Authors.
# See the Amp UNLICENSE file for details.

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

trap log_error ERR

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
