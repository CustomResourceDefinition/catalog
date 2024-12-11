input=/app/configuration.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
echo "Templating (git) ..."

mkdir -p /tmp/gen

function generate {
    cd /tmp/git || return
    git checkout "$2" &>/dev/null
    {
        for path in ${3}; do
            test -d "/tmp/git/$path/" && find "/tmp/git/$path/" -type f \( -iname '*.yaml' -o -iname '*.yml' \) -exec sh -c "echo '---'; cat \$0" {} \;
        done
        for path in ${4}; do
            echo '---'
            test -d "/tmp/git/$path/" && kustomize build "/tmp/git/$path/"
        done
        for path in ${5}; do
            find /tmp/gen -type f -delete || true
            controller-gen crd "paths=/tmp/git/$path/" output:crd:dir=/tmp/gen output:none 2>/dev/null || true
            find /tmp/gen -type f -iname "*.yaml" -exec sh -c 'echo "---"; cat $0;' {} \;
        done
    } | yq 'select(.kind == "CustomResourceDefinition")' > "$1"
}

yq eval '.[] | select(.kind == "git")' $input -o json | jq -rc | while IFS= read -r item; do
    repository=$(echo "$item" | jq -r '.repository' -)
    includeHead=$(echo "$item" | jq -r '.includeHead?' -)
    versionPrefix=$(echo "$item" | jq -r '.versionPrefix? // ""' -)
    versionSuffix=$(echo "$item" | jq -r '.versionSuffix? // "$"' -)
    paths=$(echo "$item" | jq -r '.searchPaths[]? // ""' -)
    kustomizations=$(echo "$item" | jq -r '.kustomizationPaths[]? // ""' -)
    genPaths=$(echo "$item" | jq -r '.genPaths[]? // ""' -)
    combinedname=$(echo "$repository" | rev | cut -d/ -f1-2 | rev)
    name=$(echo "$combinedname" | cut -d/ -f1)
    entry=$(echo "$item" | jq -r '.name' -)

    printf '  - %s\n' "$repository"
    printf '    - %s\n' "$name"

    rm -rf /tmp/git &>/dev/null
    mkdir -p /tmp/git
    cd /tmp/git || continue
    git clone --quiet --recursive "$repository" /tmp/git

    mkdir -p "$(printf "$output" "$name" "$entry" | tr '[:upper:]' '[:lower:]')" || true

    git tag | grep -E "^${versionPrefix}[0-9]{1,}.[0-9]{1,}.[0-9]{1,}${versionSuffix}" | sort -V > /tmp/versions
    if [ "$includeHead" = "true" ]; then
        version=$(git branch --show-current)
        git branch --show-current >> /tmp/versions
    else
        version=$(tail -n1 /tmp/versions)
    fi

    versions=$(cat /tmp/versions)
    file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')

    cd - >/dev/null || exit 1
    { test -s /tmp/versions && [ ! "$version" = "" ]; } || continue

    generate "$file" "$version" "$paths" "$kustomizations" "$genPaths"

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
        file=$(printf "$outputfile" "$name" "$entry" "$version" | tr '[:upper:]' '[:lower:]')
        generate "$file" "$version" "$paths" "$kustomizations" "$genPaths"
    done
done
echo
