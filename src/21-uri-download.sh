input="$(pwd)/$1"
output="$(pwd)/$3/templates/%s/%s/"
outputfile="$(pwd)/$3/templates/%s/%s/%s.yaml"
echo "Downloading ..."
yq eval '.[] | select(.kind == "http")' $input -o json | jq -rc | while IFS= read -r item; do
    name=$(echo "$item" | jq -r '.name' -)
    groups=$(echo "$item" | jq -r '.apiGroups[]' -)

    echo "  - $name"

    known=1
    for group in ${groups}; do
        if [ ! -d "$2/$group" ]; then
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

        mkdir -p "$(printf "$output" "$name" "$version" | tr '[:upper:]' '[:lower:]')" || true
        for path in ${paths}; do
            file=$(printf "$outputfile" "$name" "$version" "$(echo "$path" | md5sum | cut -d' ' -f1)" | tr '[:upper:]' '[:lower:]')
            wget -q "$baseUri/$path" -O "$file" || true
        done
    done
done
echo
