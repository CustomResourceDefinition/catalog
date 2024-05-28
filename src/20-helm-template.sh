input=/app/configuration.yaml
output=/templates/%s/
outputfile=/templates/%s/%s-%s.yaml
# FIXME: break on errors?
repositories=$(yq '.[].repository' $input)
for repository in ${repositories}; do
    name=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .name')
    entries=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .entries[]')
    echo "Templating $name from $repository ..."
    mkdir -p "$(printf $output $name)" || true
    for entry in ${entries}; do
        # FIXME: if seen before something else perhaps?
        versions=$(helm search repo "$name/$entry" --versions -o json | jq -rc 'reverse | .[].version')
        for version in ${versions}; do
            file=$(printf $outputfile $name $entry $version)
            helm template --include-crds "$name" "$name/$entry" | yq 'select(.kind == "CustomResourceDefinition")' > $file
            # FIXME: check for empty file
        done
    done
done
