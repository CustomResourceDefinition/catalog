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

