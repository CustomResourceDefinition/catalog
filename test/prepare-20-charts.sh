set -e

echo "Setup test http charts ... "

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
    yq '.spec.group = "chart.conditional"' < /app/test/fixtures/test-crd.yaml
    echo '{{- end }}'
} > templated/templates/crd.yaml
cd - &>/dev/null || true

cp -r /tmp/charts/base/regular /tmp/charts/regular/1.0
cp /app/test/fixtures/test-crd.yaml /tmp/charts/regular/1.0/crds/old-crd.yaml
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
    yq '.spec.group = "chart.unconditional"' < /app/test/fixtures/test-crd.yaml
} > /tmp/charts/templated/2.0/templates/crd.yaml

mkdir -p /verified-schema/chart.local /verified-schema/chart.old /verified-schema/chart.conditional /verified-schema/chart.unconditional
cp /app/test/fixtures/test_v1.json /verified-schema/chart.local/
cp /app/test/fixtures/test_v1.json /verified-schema/chart.old/
cp /app/test/fixtures/test_v1.json /verified-schema/chart.conditional/
cp /app/test/fixtures/test_v1.json /verified-schema/chart.unconditional/
yq '.spec.group = "chart.local"' < /app/test/fixtures/test-crd.yaml > /verified-schema/chart.local/test.yaml
yq '.spec.group = "chart.old"' < /app/test/fixtures/test-crd.yaml > /verified-schema/chart.old/test.yaml
yq '.spec.group = "chart.conditional"' < /app/test/fixtures/test-crd.yaml > /verified-schema/chart.conditional/test.yaml
yq '.spec.group = "chart.unconditional"' < /app/test/fixtures/test-crd.yaml > /verified-schema/chart.unconditional/test.yaml

cd /repository/http/

helm package /tmp/charts/regular/1.0
helm package /tmp/charts/regular/1.5
helm package /tmp/charts/regular/2.0

helm package /tmp/charts/templated/1.0
helm package /tmp/charts/templated/2.0

helm repo index .

echo
