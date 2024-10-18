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

cd /repository/http/

helm package /tmp/charts/regular/1.0
helm push regular-1.0.0.tgz oci://localhost:5000/myorg/myrepo
helm package /tmp/charts/regular/1.5
helm push regular-1.5.0.tgz oci://localhost:5000/myorg/myrepo
helm package /tmp/charts/regular/2.0
helm push regular-2.0.0.tgz oci://localhost:5000/myorg/myrepo

helm package /tmp/charts/templated/1.0
helm push templated-1.0.0.tgz oci://localhost:5000/myorg/myrepo
helm package /tmp/charts/templated/2.0
helm push templated-2.0.0.tgz oci://localhost:5000/myorg/myrepo

helm repo index .

echo
