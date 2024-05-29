echo "Setup chart ..."

cd /repository
helm package /chart

cp -r /chart /chart2
rm /chart2/crds/crd-2.yaml
yq -i '.version = "0.0.0"' /chart2/Chart.yaml
helm package /chart2

helm repo index .
echo
