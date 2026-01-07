#!/usr/bin/env bash

set -e

title="$1"
schema="$2"
verified="$3"

echo "Verifying $title ..."

{
    cd "$schema" || true
    find . -type f \( -name "*.json" -o -name "*.yaml" \)
    cd - &>/dev/null || true
} | sort >/tmp/schema.list

{
    cd "$verified" || true
    find . -type f \( -name "*.json" -o -name "*.yaml" \)
    cd - &>/dev/null || true
} | sort >/tmp/verified.list

diff /tmp/schema.list /tmp/verified.list

# shellcheck disable=SC2044
for f in $(find "$verified" -type f -name "*.json"); do
    diff "${f#/verified-}" "$f"
done

set +e
# shellcheck disable=SC2044
for f in $(find "$verified" -type f -name "*.yaml"); do
    if ! diff <(yq -o=json "${f#/verified-}" | jq -S | yq -P) <(yq -o=json "$f" | jq -S | yq -P); then # complex diff for comparison regardless of whitespace and ordering
        echo
        echo
        echo "Failed check on $f"
        exit 1
    fi
done

echo
echo " --- Passed - $title"
echo
