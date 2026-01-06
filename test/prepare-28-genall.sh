#!/usr/bin/env bash

set -e

verified="$3"

echo "Setup test genall source ... "

mkdir -p "$verified/source.example.com"
cp test/fixtures/foo_foo.json "$verified/source.example.com/"
cp test/fixtures/foo.yaml "$verified/source.example.com/foo.yaml"

echo
