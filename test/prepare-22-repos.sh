set -e
base=$(pwd)
echo "Setup test git charts ... "

mkdir -p "$3/repository/git" > /dev/null 2>&1 || true
cd "$3/repository/git"

git config --global user.email "test@runner.local"
git config --global user.name "Test Runner"
git init --initial-branch=main

mkdir crds
cp "$base/test/fixtures/test-crd.yaml" crds/crd.yaml
yq -i '.spec.group = "chart.git"' crds/crd.yaml
yq -i '.version = "1.0.0"' crds/crd.yaml

mkdir kustomizations
cp crds/crd.yaml kustomizations/crd.yaml
{
echo 'apiVersion: kustomize.config.k8s.io/v1beta1'
echo 'kind: Kustomization'
echo 'resources:'
echo '  - crd.yaml'
} > kustomizations/kustomization.yaml

git add crds/crd.yaml
git add kustomizations
git commit -m commit >/dev/null
git tag v1.0.0
git tag v2.0.0
git tag v10.0.0

cd - > /dev/null 2>&1

echo
