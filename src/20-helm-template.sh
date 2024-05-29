input=/app/configuration.yaml
output=/templates/%s/
outputfile=/templates/%s/%s-%s.yaml
# FIXME: break on errors?
repositories=$(yq '.[].repository' $input)
echo "Templating ..."
for repository in ${repositories}; do
    name=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .name')
    entries=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .entries[]')
    printf '  - %s from %s' "$name" "$repository"
    mkdir -p "$(printf $output $name)" || true
    for entry in ${entries}; do
        file=$(printf $outputfile $name $entry 'latest')
        helm template --include-crds "$name" "$name/$entry" | yq 'select(.kind == "CustomResourceDefinition")' | yq eval 'del(.. | .description?)' > "$file"
        groups=$(yq .spec.group < $file | grep -v '\---' | uniq)

        known=1
        for group in ${groups}; do
            test -d /schema/$group || known=0
        done
        if [ $known -eq 1 ]; then
            printf ' - using latest version only\n'
            break
        fi

        rm "$file"
        printf ' - using all versions\n'
        versions=$(helm search repo "$name" --versions -o json | jq -rc  --arg name "$name/$entry" '[.[] | select(.name == $name)] | reverse | .[].version')
        for version in ${versions}; do
            file=$(printf $outputfile $name $entry $version)
            helm template --include-crds "$name" "$name/$entry" --version "$version" | yq 'select(.kind == "CustomResourceDefinition")' | yq eval 'del(.. | .description?)' > "$file"
        done
    done
done
