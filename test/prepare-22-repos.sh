set -e

echo "Setup test git charts ... "
mkdir -p /repository/git &>/dev/null || true
cd /repository/git

git config --global user.email "test@runner.local"
git config --global user.name "Test Runner"
git init --initial-branch=main

mkdir crds
cp /app/test/fixtures/test-crd.yaml crds/crd.yaml
yq -i '.spec.group = "chart.git"' crds/crd.yaml

mkdir kustomizations
cp crds/crd.yaml kustomizations/crd.yaml
{
echo 'apiVersion: kustomize.config.k8s.io/v1beta1'
echo 'kind: Kustomization'
echo 'resources:'
echo '  - crd.yaml'
} > kustomizations/kustomization.yaml

mkdir -p /verified-schema/chart.git
cp /app/test/fixtures/test_v1.json /verified-schema/chart.git/
cp crds/crd.yaml /verified-schema/chart.git/test.yaml

mkdir source
cp /app/test/fixtures/source/* source

git add crds/crd.yaml
git add kustomizations
git add source
git commit -m commit >/dev/null
git tag v1.0.0
git tag v2.0.0
git tag v10.0.0

echo
