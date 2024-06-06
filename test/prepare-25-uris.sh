set -e
base=$(pwd)
echo "Setup test http uris ... "

cd "$3/repository/http"

cp "$base/test/fixtures/test-crd.yaml" ./chart-1.0.0.yaml
yq -i '.version = "1.0.0"' ./chart-1.0.0.yaml
yq -i '.spec.group = "chart.uri1"' ./chart-1.0.0.yaml

cp "$base/test/fixtures/test-crd.yaml" ./chart-2.0.0.yaml
yq -i '.version = "2.0.0"' ./chart-2.0.0.yaml
yq -i '.spec.group = "chart.uri"' ./chart-2.0.0.yaml

cd - > /dev/null 2>&1

echo
