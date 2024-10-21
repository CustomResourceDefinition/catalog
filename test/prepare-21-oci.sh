set -e

echo "Setup test oci charts ... "

rm -rf -- /tmp/charts &>/dev/null || true
mkdir -p /tmp/charts/regular || true
mkdir -p /tmp/charts/templated || true
mkdir -p /tmp/charts/base || true
cd /tmp/charts/base || true
helm create regular
mkdir -p regular/crds
cp /app/test/fixtures/test-crd.yaml regular/crds/crd.yaml
helm create templated
{
    echo '{{- if .Values.output }}'
    yq '.spec.group = "chart.conditional-oci"' < /app/test/fixtures/test-crd.yaml
    echo '{{- end }}'
} > templated/templates/crd.yaml
cd - &>/dev/null || true

cp -r /tmp/charts/base/regular /tmp/charts/regular/1.0
cp /app/test/fixtures/test-crd.yaml /tmp/charts/regular/1.0/crds/old-crd.yaml
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
    yq '.spec.group = "chart.unconditional-oci"' < /app/test/fixtures/test-crd.yaml
} > /tmp/charts/templated/2.0/templates/crd.yaml

cd /tmp/charts

helm package /tmp/charts/regular/1.0
helm package /tmp/charts/regular/1.5
helm package /tmp/charts/regular/2.0

helm package /tmp/charts/templated/1.0
helm package /tmp/charts/templated/2.0

helm push --plain-http regular-1.0.0.tgz oci://registry:5000/oci
helm push --plain-http regular-1.5.0.tgz oci://registry:5000/oci
helm push --plain-http regular-2.0.0.tgz oci://registry:5000/oci
helm push --plain-http templated-1.0.0.tgz oci://registry:5000/oci
helm push --plain-http templated-2.0.0.tgz oci://registry:5000/oci

cd - &>/dev/null

echo
