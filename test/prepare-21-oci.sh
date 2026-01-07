#!/usr/bin/env bash

set -e

verified="$3"
cwd=$(pwd)

echo "Setup test oci charts ... "

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
    yq '.spec.group = "chart.conditional-oci"' <"$cwd/test/fixtures/test-crd.yaml"
    echo '{{- end }}'
} >templated/templates/crd.yaml
cd - &>/dev/null

cp -r /tmp/charts/base/regular /tmp/charts/regular/1.0
cp test/fixtures/test-crd.yaml /tmp/charts/regular/1.0/crds/old-crd.yaml
yq -i '.version = "1.0.0"' /tmp/charts/regular/1.0/Chart.yaml
yq -i '.spec.group = "chart.local-oci"' /tmp/charts/regular/1.0/crds/crd.yaml
yq -i '.spec.group = "chart.old-oci"' /tmp/charts/regular/1.0/crds/old-crd.yaml

cp -r /tmp/charts/base/regular /tmp/charts/regular/1.5
yq -i '.version = "1.5.0"' /tmp/charts/regular/1.5/Chart.yaml
yq -i '.spec.group = "chart.local-oci"' /tmp/charts/regular/1.5/crds/crd.yaml

cp -r /tmp/charts/base/regular /tmp/charts/regular/2.0
yq -i '.version = "2.0.0"' /tmp/charts/regular/2.0/Chart.yaml
yq -i '.spec.group = "chart.local-oci"' /tmp/charts/regular/2.0/crds/crd.yaml

cp -r /tmp/charts/base/templated /tmp/charts/templated/1.0
yq -i '.version = "1.0.0"' /tmp/charts/templated/1.0/Chart.yaml
cp -r /tmp/charts/base/templated /tmp/charts/templated/2.0
yq -i '.version = "2.0.0"' /tmp/charts/templated/2.0/Chart.yaml
{
    echo '{{- if .Values.output }}'
    echo '{{- fail "value for .Values.output is not allowed in this version" }}'
    echo '{{- end }}'
    yq '.spec.group = "chart.unconditional-oci"' <test/fixtures/test-crd.yaml
} >/tmp/charts/templated/2.0/templates/crd.yaml

mkdir -p "$verified/chart.local-oci" "$verified/chart.old-oci" "$verified/chart.conditional-oci" "$verified/chart.unconditional-oci"
cp test/fixtures/test_v1.json "$verified/chart.local-oci/"
cp test/fixtures/test_v1.json "$verified/chart.old-oci/"
cp test/fixtures/test_v1.json "$verified/chart.conditional-oci/"
cp test/fixtures/test_v1.json "$verified/chart.unconditional-oci/"
yq '.spec.group = "chart.local-oci"' <test/fixtures/test-crd.yaml >"$verified/chart.local-oci/test.yaml"
yq '.spec.group = "chart.old-oci"' <test/fixtures/test-crd.yaml >"$verified/chart.old-oci/test.yaml"
yq '.spec.group = "chart.conditional-oci"' <test/fixtures/test-crd.yaml >"$verified/chart.conditional-oci/test.yaml"
yq '.spec.group = "chart.unconditional-oci"' <test/fixtures/test-crd.yaml >"$verified/chart.unconditional-oci/test.yaml"

cd /tmp/charts

helm package /tmp/charts/regular/1.0
helm package /tmp/charts/regular/1.5
helm package /tmp/charts/regular/2.0

helm package /tmp/charts/templated/1.0
helm package /tmp/charts/templated/2.0

helm push --plain-http regular-1.0.0.tgz oci://registry.127.0.0.1.nip.io:8085/oci
helm push --plain-http regular-1.5.0.tgz oci://registry.127.0.0.1.nip.io:8085/oci
helm push --plain-http regular-2.0.0.tgz oci://registry.127.0.0.1.nip.io:8085/oci
helm push --plain-http templated-1.0.0.tgz oci://registry.127.0.0.1.nip.io:8085/oci
helm push --plain-http templated-2.0.0.tgz oci://registry.127.0.0.1.nip.io:8085/oci

cd - &>/dev/null

echo
