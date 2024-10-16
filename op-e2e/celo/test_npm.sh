#!/bin/bash
#shellcheck disable=SC1091
set -eo pipefail

source shared.sh
npm test
