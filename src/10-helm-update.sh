input=/app/configuration.yaml
yq eval '.[] | select(.kind == "helm")' $input -o json | jq -rc | while IFS= read -r item; do
    repository=$(echo "$item" | jq -r '.repository' -)
    name=$(echo "$item" | jq -r '.name' -)
    helm repo list 2>/dev/null | grep -q "^${name}[:space:]" && continue
    helm repo add "$name" "$repository"
done
echo
