input=/app/configuration.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
repositories=$(yq '.[] | select(.kind == "helm-oci") | .repository' $input)
echo "Templating (oci) ..."
for repository in ${repositories}; do
    versions=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository) | .additionalVersions[]')
    combinedname=$(echo "$repository" | rev | cut -d/ -f1-2 | rev)
    name=$(echo "$combinedname" | cut -d/ -f1)
    entry=$(echo "$combinedname" | cut -d/ -f2)

    printf '  - %s\n' "$repository"

    oci_values_file_of "$input" "$repository" > /tmp/values

    mkdir -p "$(printf "$output" "$name" "$entry" | tr '[:upper:]' '[:lower:]')" || true
    version=$(helm template $HELMFLAGS "$repository" -f /tmp/values 2>&1 | head -n1 | cut -d/ -f2- | cut -d: -f2)
    file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
    helm template $HELMFLAGS --include-crds "$repository" --version "$version" -f /tmp/values 2>/dev/null | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
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
    else
        lookupversion=$version
    fi

    for version in ${versions}; do
        printf '      - version %s\n' "$version"
        oci_values_file_of "$input" "$repository" "$version" > /tmp/values
        file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
        helm template $HELMFLAGS --include-crds "$repository" --version "$version" -f /tmp/values 2>/dev/null | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
    done

    printf '      - version %s\n' "$lookupversion"
done
echo
