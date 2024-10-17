input=/app/configuration.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
repositories=$(yq '.[] | select(.kind == "helm-oci") | .repository' $input)
echo "Templating (oci) ..."
for repository in ${repositories}; do
    versions=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .additionalVersions[]')
    combinedname=$(echo "$repository" | rev | cut -d/ -f1-2 | rev)
    name=$(echo "$combinedname" | cut -d/ -f2)
    entry=$(echo "$combinedname" | cut -d/ -f1)

    printf '  - %s\n' "$repository"

    values_file_of "$input" "$repository" "$name" > /tmp/values

    mkdir -p "$(printf "$output" "$name" "$entry" | tr '[:upper:]' '[:lower:]')" || true
    version=$(helm template "$repository" -f /tmp/values 2>&1 | head -n1 | cut -d: -f3)
    file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
    helm template --include-crds "$repository" --version "$version" -f /tmp/values 2>/dev/null | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
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
        values_file_of "$input" "$repository" "$name" "$version" > /tmp/values
        file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
        helm template --include-crds "$repository" --version "$version" -f /tmp/values 2>/dev/null | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
    done
done
echo
