cd /repository; python3 -m http.server 2>&1 | grep -v '" 200 -' &

input=/app/configuration.yaml
output=/templates/%s/
outputfile=/templates/%s/%s-%s.yaml
# FIXME: break on errors?
repositories=$(yq '.[].repository' $input)
echo "Templating ..."
for repository in ${repositories}; do
    name=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .name')
    entries=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .entries[]')
    echo "  - $name from $repository"
    mkdir -p "$(printf $output $name)" || true
    for entry in ${entries}; do
        # FIXME: if seen before something else perhaps?
        # FIXME: remove slice
        versions=$(helm search repo "$name" --versions -o json | jq -rc  --arg name "$name/$entry" '[.[] | select(.name == $name)] | .[0:10] | reverse | .[].version')
        for version in ${versions}; do
            file=$(printf $outputfile $name $entry $version)
            helm template --include-crds "$name" "$name/$entry" --version "$version" | yq 'select(.kind == "CustomResourceDefinition")' | yq eval 'del(.. | .description?)' > "$file"
        done
    done
done
