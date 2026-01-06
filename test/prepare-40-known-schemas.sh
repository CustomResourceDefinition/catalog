#!/usr/bin/env bash

mode="$1"
schema="$2"
verified="$3"

echo "Setup known schemas ..."

if [ "$mode" = "only-latest" ]; then
    rm -r "$verified/chart.conditional"
    rm -r "$verified/chart.conditional-oci"
    rm -r "$verified/chart.old"
    rm -r "$verified/chart.old-oci"
    rm -r "$verified/chart.uri1"

    cd "$verified" || exit 1
    find . -print | while read -r item; do
        if [ -d "$item" ]; then
            mkdir -p "$schema/$item"
        elif [ -f "$item" ]; then
            touch "$schema/$item"
        fi
    done
    cd - || exit 1
fi

echo "  - updated for test '$mode'"
echo
