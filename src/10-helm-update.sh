input=/app/configuration.yaml
repositories=$(yq '.[].repository' $input)
for repository in ${repositories}; do
    name=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .name')
    helm repo list 2>/dev/null | grep -q "^${name}[:space:]" && continue
    helm repo add "$name" "$repository"
done
helm repo update
echo
