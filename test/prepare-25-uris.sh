set -e

echo "Setup test uris ... "
cd /repository

cp /chart/crds/crd-1.yaml ./chart.yaml
yq -i '.version = "1.0.0"' ./chart.yaml
yq -i '.spec.group = "chart.uri"' ./chart.yaml

echo
