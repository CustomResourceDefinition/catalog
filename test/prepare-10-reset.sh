#!/usr/bin/env bash

set -e

title="$1"
schema="$2"
verified="$3"
repository="$4"

echo "Reset for a fresh test - ${title:?} ... "

helm repo ls -o yaml | yq -r '.[].name' | xargs -I {} helm repo rm {}

rm -rf "${schema:?}/*" &>/dev/null || true
rm -rf "${repository:?}/*" &>/dev/null || true
rm -rf "${verified:?}/*" &>/dev/null || true

mkdir -p "$verified"
mkdir -p "$repository/git"
mkdir -p "$repository/http"

echo "done"
echo
