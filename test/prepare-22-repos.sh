set -e

echo "Setup test git charts ... "
mkdir -p /repository/git &>/dev/null || true
cd /repository/git
rm -rf -- * &>/dev/null || true

git config --global user.email "test@runner.local"
git config --global user.name "Test Runner"
git init --initial-branch=main

mkdir crds
cp /chart/crds/crd-1.yaml crds/crd.yaml
yq -i '.spec.group = "chart.git"' crds/crd.yaml
yq -i '.version = "1.0.0"' crds/crd.yaml

mkdir kustomizations
cp crds/crd.yaml kustomizations/crd.yaml
echo 'apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - crd.yaml' > kustomizations/kustomization.yaml

git add crds/crd.yaml
git add kustomizations
git commit -m commit >/dev/null
git tag v1.0.0
git tag v2.0.0
git tag v10.0.0

echo
