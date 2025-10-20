#!/usr/bin/env bash
set -euo pipefail

# Cria os arquivos estáticos para os starters

PROJECT_ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
STARTERS_DIR="${PROJECT_ROOT_DIR}/starters"
STATIC_DIR="${PROJECT_ROOT_DIR}/static"

CODE_TEMPLATE_DIR="${STATIC_DIR}/code-templates"

# zip_starter "${starter_dir}" "${starter_name}"
function zip_starter() {
    local starter_dir="${1}"
    local zip_file="${starter_dir}.zip"
    pushd "${starter_dir}"
    [ -e "${zip_file}" ] && rm "${zip_file}"
    zip -r "${zip_file}" .
    mv "${zip_file}" "${CODE_TEMPLATE_DIR}/${2}/starter.zip"
    popd
}

for starter_dir in "${STARTERS_DIR}"/*; do
    zip_starter "${starter_dir}" "${starter_dir##*/}"
done