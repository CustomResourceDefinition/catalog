set -e

echo "Setup test http charts ... "

helm package /chart

rm -rf /chart1 &>/dev/null || true
cp -r /chart /chart1
yq -i '.version = "1.0.0"' /chart1/Chart.yaml
helm package /chart1

rm -rf /chart1.5 &>/dev/null || true
cp -r /chart /chart1.5
yq -i '.version = "1.5.0"' /chart1.5/Chart.yaml
helm package /chart1.5

rm -rf /chart2 &>/dev/null || true
cp -r /chart /chart2
yq -i '.spec.group = "chart.new"' /chart2/crds/crd-1.yaml
yq -i '.version = "2.0.0"' /chart2/Chart.yaml
helm package /chart2

helm package /chart-value-based

helm repo index .
echo
