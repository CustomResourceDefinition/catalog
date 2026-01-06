#!/usr/bin/env bash

set -e

verified="$3"
repository="$4"

echo "Setup test http uris ... "

cp test/fixtures/test-crd.yaml "$repository/http/chart-1.0.0.yaml"
yq -i '.spec.group = "chart.uri1"' "$repository/http/chart-1.0.0.yaml"

cp test/fixtures/test-crd.yaml "$repository/http/chart-2.0.0.yaml"
yq -i '.spec.group = "chart.uri"' "$repository/http/chart-2.0.0.yaml"

mkdir -p "$verified/chart.uri" "$verified/chart.uri1"
cp test/fixtures/test_v1.json "$verified/chart.uri/"
cp test/fixtures/test_v1.json "$verified/chart.uri1/"
cp "$repository/http/chart-2.0.0.yaml" "$verified/chart.uri/test.yaml"
cp "$repository/http/chart-1.0.0.yaml" "$verified/chart.uri1/test.yaml"

echo
