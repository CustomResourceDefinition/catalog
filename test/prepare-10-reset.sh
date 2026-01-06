#!/usr/bin/env bash

set -e

schema="$2"
verified="$3"
repository="$4"

echo "Reset for a fresh test ... "

helm repo ls -o yaml | yq -r '.[].name' | xargs -I {} helm repo rm {}

rm -rf "${schema:?}/*" &>/dev/null || true
rm -rf "${repository:?}" &>/dev/null || true
rm -rf "${verified:?}" &>/dev/null || true

mkdir -p "$verified"
mkdir -p "$repository/git"

echo "done"
echo
