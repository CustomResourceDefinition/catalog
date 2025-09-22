set -e

echo "Setup test http uris ... "
cd /repository/http

cp /app/test/fixtures/test-crd.yaml ./chart-1.0.0.yaml
yq -i '.spec.group = "chart.uri1"' ./chart-1.0.0.yaml

cp /app/test/fixtures/test-crd.yaml ./chart-2.0.0.yaml
yq -i '.spec.group = "chart.uri"' ./chart-2.0.0.yaml

mkdir -p /verified-schema/chart.uri /verified-schema/chart.uri1
cp /app/test/fixtures/test_v1.json /verified-schema/chart.uri/
cp /app/test/fixtures/test_v1.json /verified-schema/chart.uri1/
cp ./chart-2.0.0.yaml /verified-schema/chart.uri/test.yaml
cp ./chart-1.0.0.yaml /verified-schema/chart.uri1/test.yaml

echo
