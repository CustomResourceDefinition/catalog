input=/app/oci-charts.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
repositories=$(yq '.[].repository' $input)
echo "Templating (oci) ..."
for repository in ${repositories}; do
    versions=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .additionalVersions[]')
    combinedname=$(echo "$repository" | rev | cut -d/ -f1-2 | rev)
    name=$(echo "$combinedname" | cut -d/ -f1)
    entry=$(echo "$combinedname" | cut -d/ -f2)

    printf '  - %s\n' "$repository"

    #shellcheck disable=SC2059
    mkdir -p "$(printf "$output" "$name" "$entry")" || true
    version=$(helm template "$repository" 2>&1 | head -n1 | cut -d\: -f3)
    #shellcheck disable=SC2059
    file=$(printf "$outputfile" "$name" "$entry" "$version")
    helm template --include-crds "$repository" --version "$version" 2>/dev/null | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
    groups=$(yq .spec.group < $file | grep -v '\---' | grep -v null | uniq)

    known=1
    for group in ${groups}; do
        if [ ! -d /schema/$group ]; then
            known=0
            echo "      - $group is unknown -> render all versions"
        fi
    done
    if [ $known -eq 1 ]; then
        printf '      - version %s\n' "$version"
        continue
    fi

    for version in ${versions}; do
        printf '      - version %s\n' "$version"
        #shellcheck disable=SC2059
        file=$(printf "$outputfile" "$name" "$entry" "$version")
        helm template --include-crds "$repository" --version "$version" 2>/dev/null | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
    done
done
echo
