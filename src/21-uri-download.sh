input=/app/manifest-uris.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
echo "Downloading ..."
yq eval '.[]' $input -o json | jq -rc |while IFS= read -r item; do
    name=$(echo "$item" | jq -r '.name' -)
    version=$(echo "$item" | jq -r '.version' -)
    baseUri=$(echo "$item" | jq -r '.baseUri' -)
    paths=$(echo "$item" | jq -r '.paths[]' -)

    echo "  - $name version $version"

    #shellcheck disable=SC2059
    mkdir -p "$(printf "$output" "uri" "$name")" || true
    for path in ${paths}; do
        #shellcheck disable=SC2059
        file=$(printf "$outputfile" "uri" "$name" "$version")
        wget -q "$baseUri/$path" -O "$file" || true
    done
done
echo
