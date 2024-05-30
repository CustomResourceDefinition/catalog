input=/app/manifest-uris.yaml
output=/templates/%s/%s/
outputfile=/templates/%s/%s/%s.yaml
echo "Downloading ..."
yq eval '.[]' $input -o json | jq -rc | while IFS= read -r item; do
    name=$(echo "$item" | jq -r '.name' -)
    groups=$(echo "$item" | jq -r '.apiGroups[]' -)

    echo "  - $name"

    known=1
    for group in ${groups}; do
        if [ ! -d /schema/$group ]; then
            known=0
            echo "    - $group is unknown -> render all versions"
        fi
    done

    if [ $known -eq 1 ]; then
        crds=$(echo "$item" | jq -rc '.crds | sort_by(.version | split(".") | map(tonumber)) | reverse | .[0]' -)
    else
        crds=$(echo "$item" | jq -rc '.crds[]' -)
    fi

    echo "$crds" | while IFS= read -r crd; do
        version=$(echo "$crd" | jq -r '.version' -)
        baseUri=$(echo "$crd" | jq -r '.baseUri' -)
        paths=$(echo "$crd" | jq -r '.paths[]' -)

        echo "    - version $version"

        #shellcheck disable=SC2059
        mkdir -p "$(printf "$output" "$name" "$version")" || true
        for path in ${paths}; do
            #shellcheck disable=SC2059
            file=$(printf "$outputfile" "$name" "$version" "$(echo "$path" | md5sum | cut -d' ' -f1)")
            wget -q "$baseUri/$path" -O "$file" || true
        done
    done
done
echo