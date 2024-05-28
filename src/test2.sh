touch /tmp/values
echo 'manager:
  collectorImage:
    repository: "otel/opentelemetry-collector-k8s"' > /tmp/values

# helm search repo opentelemetry-helm/opentelemetry-operator --versions -o yaml
# helm show chart opentelemetry-helm/opentelemetry-operator

# helm template -s crds --skip-tests --no-hooks --dry-run --disable-openapi-validation --generate-name --include-crds opentelemetry-helm/opentelemetry-operator

helm template --include-crds opentelemetry-helm opentelemetry-helm/opentelemetry-operator --values /tmp/values | yq 'select(.kind == "CustomResourceDefinition")' > crds.yaml

output=crds
mkdir -p $output
python /app/helpers/split.py crds.yaml "${output}/f"



python /app/helpers/convert.py $output/*
