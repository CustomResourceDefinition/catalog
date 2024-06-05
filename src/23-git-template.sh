input=/app/git-charts.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
repositories=$(yq '.[].repository' $input)
echo "Templating (git) ..."

function generate {
    file=$1
    version=$2
    paths=$3
    kustomizations=$4
    $(cd /tmp/git; git checkout "$version" &>/dev/null)
    {
        for path in ${paths}; do
            test -d "/tmp/git/$path/" && find "/tmp/git/$path/" -type f \( -iname '*.yaml' -o -iname '*.yml' \) -exec sh -c "echo '---'; cat \$0" {} \;
        done
        for path in ${kustomizations}; do
            echo '---'
            test -d "/tmp/git/$path/" && kustomize build "/tmp/git/$path/"
        done
    } | tee -a /tmp/snoop | yq 'select(.kind == "CustomResourceDefinition")' > "$file"
}

yq eval '.[]' $input -o json | jq -rc | while IFS= read -r item; do
    repository=$(echo "$item" | jq -r '.repository' -)
    includeHead=$(echo "$item" | jq -r '.includeHead?' -)
    versionPrefix=$(echo "$item" | jq -r '.versionPrefix? // ""' -)
    paths=$(echo "$item" | jq -r '.searchPaths[]? // ""' -)
    kustomizations=$(echo "$item" | jq -r '.kustomizationPaths[]? // ""' -)
    combinedname=$(echo "$repository" | rev | cut -d/ -f1-2 | rev)
    name=$(echo "$combinedname" | cut -d/ -f1)
    entry=$(echo "$item" | jq -r '.name' -)

    printf '  - %s\n' "$repository"

    rm -rf /tmp/git &>/dev/null
    mkdir -p /tmp/git
    git clone --quiet --recursive "$repository" /tmp/git

    mkdir -p "$(printf "$output" "$name" "$entry" | tr '[:upper:]' '[:lower:]')" || true

    printf '    - %s\n' "$name"

    cd /tmp/git
    git tag | grep -E "^${versionPrefix}[0-9]{1,}.[0-9]{1,}.[0-9]{1,}$" | sort -V > /tmp/versions
    version=$([ "$includeHead" = "true" ] && git branch --show-current || tail -n1 /tmp/versions)
    [ "$includeHead" = "true" ] && git branch --show-current >> /tmp/versions

    versions=$(cat /tmp/versions)
    file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')

    cd - >/dev/null
    (test -s /tmp/versions && [ ! "$version" = "" ]) || continue

    generate "$file" "$version" "$paths" "$kustomizations"

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
        file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
        #shellcheck enable=SC2059
        generate "$file" "$version" "$paths" "$kustomizations"
    done
done
echo
