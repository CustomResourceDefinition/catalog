echo "Template it"
exit 0

input=/app/configuration.yaml
repositories=$(yq '.[].repository' $input)
for repository in ${repositories}; do
    name=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .name')
    entries=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .entries[]')
    for entry in ${entries}; do
        # if seen before
        helm search repo "$entry" --versions -o yaml
        # else
        helm search repo "$entry" --versions -o yaml
    done
done


# touch /tmp/values
# echo 'manager:
#   collectorImage:
#     repository: "otel/opentelemetry-collector-k8s"' > /tmp/values

# # helm search repo opentelemetry-helm/opentelemetry-operator --versions -o yaml
# # helm show chart opentelemetry-helm/opentelemetry-operator

# # helm template -s crds --skip-tests --no-hooks --dry-run --disable-openapi-validation --generate-name --include-crds opentelemetry-helm/opentelemetry-operator

# helm template --include-crds opentelemetry-helm opentelemetry-helm/opentelemetry-operator --values /tmp/values | yq 'select(.kind == "CustomResourceDefinition")' > crds.yaml
