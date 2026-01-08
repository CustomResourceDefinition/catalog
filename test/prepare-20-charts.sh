#!/usr/bin/env bash

set -e

verified="$3"
repository="$4"
cwd=$(pwd)

echo "Setup test http charts ... "

rm -rf -- /tmp/charts &>/dev/null || true
mkdir -p /tmp/charts/regular
mkdir -p /tmp/charts/templated
mkdir -p /tmp/charts/base
cd /tmp/charts/base
helm create regular
mkdir -p regular/crds
cp "$cwd/test/fixtures/test-crd.yaml" regular/crds/crd.yaml
helm create templated
{
    echo '{{- if .Values.output }}'
    yq '.spec.group = "chart.conditional"' <"$cwd/test/fixtures/test-crd.yaml"
    echo '{{- end }}'
} >templated/templates/crd.yaml
cd - &>/dev/null

cp -r /tmp/charts/base/regular /tmp/charts/regular/1.0
cp test/fixtures/test-crd.yaml /tmp/charts/regular/1.0/crds/old-crd.yaml
yq -i '.version = "1.0.0"' /tmp/charts/regular/1.0/Chart.yaml
yq -i '.spec.group = "chart.local"' /tmp/charts/regular/1.0/crds/crd.yaml
yq -i '.spec.group = "chart.old"' /tmp/charts/regular/1.0/crds/old-crd.yaml

cp -r /tmp/charts/base/regular /tmp/charts/regular/1.5
yq -i '.version = "1.5.0"' /tmp/charts/regular/1.5/Chart.yaml
yq -i '.spec.group = "chart.local"' /tmp/charts/regular/1.5/crds/crd.yaml

cp -r /tmp/charts/base/regular /tmp/charts/regular/2.0
yq -i '.version = "2.0.0"' /tmp/charts/regular/2.0/Chart.yaml
yq -i '.spec.group = "chart.local"' /tmp/charts/regular/2.0/crds/crd.yaml

cp -r /tmp/charts/base/templated /tmp/charts/templated/1.0
yq -i '.version = "1.0.0"' /tmp/charts/templated/1.0/Chart.yaml
cp -r /tmp/charts/base/templated /tmp/charts/templated/2.0
yq -i '.version = "2.0.0"' /tmp/charts/templated/2.0/Chart.yaml
{
    echo '{{- if .Values.output }}'
    echo '{{- fail "value for .Values.output is not allowed in this version" }}'
    echo '{{- end }}'
    yq '.spec.group = "chart.unconditional"' <test/fixtures/test-crd.yaml
} >/tmp/charts/templated/2.0/templates/crd.yaml

mkdir -p "$verified/chart.local" "$verified/chart.old" "$verified/chart.conditional" "$verified/chart.unconditional"
cp test/fixtures/test_v1.json "$verified/chart.local/"
cp test/fixtures/test_v1.json "$verified/chart.old/"
cp test/fixtures/test_v1.json "$verified/chart.conditional/"
cp test/fixtures/test_v1.json "$verified/chart.unconditional/"
yq '.spec.group = "chart.local"' <test/fixtures/test-crd.yaml >"$verified/chart.local/test.yaml"
yq '.spec.group = "chart.old"' <test/fixtures/test-crd.yaml >"$verified/chart.old/test.yaml"
yq '.spec.group = "chart.conditional"' <test/fixtures/test-crd.yaml >"$verified/chart.conditional/test.yaml"
yq '.spec.group = "chart.unconditional"' <test/fixtures/test-crd.yaml >"$verified/chart.unconditional/test.yaml"

cd "$repository/http/"

helm package /tmp/charts/regular/1.0
helm package /tmp/charts/regular/1.5
helm package /tmp/charts/regular/2.0

helm package /tmp/charts/templated/1.0
helm package /tmp/charts/templated/2.0

helm repo index .

cd - || exit 1

echo
