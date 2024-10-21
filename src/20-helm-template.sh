input=/app/configuration.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
echo "Templating (https) ..."
yq eval '.[] | select(.kind == "helm")' $input -o json | jq -rc | while IFS= read -r item; do
    repository=$(echo "$item" | jq -r '.repository' -)
    name=$(echo "$item" | jq -r '.name' -)
    entries=$(echo "$item" | jq -r '.entries[]' -)
    printf '  - %s\n' "$repository"

    for entry in ${entries}; do
        mkdir -p "$(printf "$output" "$name" "$entry" | tr '[:upper:]' '[:lower:]')" || true
        printf '    - %s\n' "$entry"
        version=$(helm show chart "$name/$entry" | yq .version)

        helm_values_file_of "$input" "$repository" "$name" "$version" > /tmp/values
        file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
        helm template --include-crds "$name" "$name/$entry" -f /tmp/values -n not-default --version "$version" | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
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

        rm "$file"
        versions=$(helm search repo "$name" --versions -o json | jq -rc  --arg name "$name/$entry" '[.[] | select(.name == $name)] | reverse | .[].version')
        for version in ${versions}; do
            printf '      - version %s\n' "$version"
            helm_values_file_of "$input" "$repository" "$name" "$version" > /tmp/values
            file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
            helm template --include-crds "$name" "$name/$entry" -f /tmp/values -n not-default --version "$version" | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
        done
    done
done
echo
