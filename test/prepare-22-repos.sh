#!/usr/bin/env bash

set -e

verified="$3"
repository="$4"
cwd=$(pwd)

echo "Setup test git charts ... "
cd "$repository/git"

git init --initial-branch=main

mkdir -p crds
cp "$cwd/test/fixtures/test-crd.yaml" crds/crd.yaml
yq -i '.spec.group = "chart.git"' crds/crd.yaml

mkdir -p kustomizations
cp crds/crd.yaml kustomizations/crd.yaml
{
    echo 'apiVersion: kustomize.config.k8s.io/v1beta1'
    echo 'kind: Kustomization'
    echo 'resources:'
    echo '  - crd.yaml'
} >kustomizations/kustomization.yaml

mkdir -p "$cwd/$verified/chart.git"
cp "$cwd/test/fixtures/test_v1.json" "$cwd/$verified/chart.git/"
cp crds/crd.yaml "$cwd/$verified/chart.git/test.yaml"

mkdir -p source
cp "$cwd/test/fixtures/source/"* source

git add crds/crd.yaml
git add kustomizations
git add source
git commit -m commit >/dev/null
git tag v1.0.0
git tag v2.0.0
git tag v10.0.0

cd - &>/dev/null

git clone --bare "$repository/git" "$repository/http/chart.git"
cd "$repository/http/chart.git"
git update-server-info -f
cd - &>/dev/null

echo
