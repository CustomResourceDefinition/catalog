set -e

echo "Setup test charts ... "
cd /repository
rm -rf * &>/dev/null || true

helm package /chart

rm -rf /chart2 &>/dev/null || true
cp -r /chart /chart2
yq -i '.spec.group = "chart.new"' /chart2/crds/crd-1.yaml
yq -i '.version = "1.0.0"' /chart2/Chart.yaml
helm package /chart2

helm repo index .
echo
