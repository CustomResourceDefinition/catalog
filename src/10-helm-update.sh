input="$(pwd)/$1"
echo "Preparing helm charts ..."
yq eval '.[] | select(.kind == "helm")' $input -o json | jq -rc | while IFS= read -r item; do
    repository=$(printf %s "$item" | jq -r '.repository' -)
    name=$(printf %s "$item" | jq -r '.name' -)
    helm repo list 2>/dev/null | grep -q "^${name}[[:space:]]" && continue
    helm repo add "$name" "$repository"
done
echo
