set -o allexport
# shellcheck source=/dev/null
test -f /app/.env && source /app/.env
set +o allexport

PATH=$PATH:/workspace/build/bin/

helm_values_file_of() {
    local input=$1
    local repository=$2
    local name=$3
    local version=$4
    local relevant_version

    if [ "" = "$4" ]; then
        version=$(yq -o json $input | jq -rc --arg repository $repository --arg name $name '.[] | select(.repository == $repository and .name == $name and .kind == "helm") | .valuesFiles // [] | map(. as $item | ($item.version | split(".")) as $version | {major: $version[0], minor: $version[1], patch: $version[2], version: $item.version} ) | sort_by(.major, .minor, .patch) | last | .version // "0.0.0"')
    fi

    keys=$(yq -o json $input | jq -rc --arg repository $repository --arg name $name '.[] | select(.repository == $repository and .name == $name and .kind == "helm") | .valuesFiles // [] | .[].version // "0.0.0"')

    if echo "$keys" | grep "$version" >/dev/null; then
        relevant_version=$version
    else
        relevant_version=$({ echo $version; echo $keys; } | tr ' ' "\n" | sort -t "." -k1,1n -k2,2n -k3,3n | grep -B1 $version | grep -v $version)
    fi

    yq -o json $input | jq -rcj --arg repository $repository --arg name $name --arg version "$relevant_version" '.[] | select(.repository == $repository and .name == $name and .kind == "helm") | .valuesFiles // [] | .[] | select(.version == $version) | .valuesFile // ""'
}

oci_values_file_of() {
    local input=$1
    local repository=$2
    local version=$3
    local relevant_version

    if [ "" = "$3" ]; then
        version=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository and .kind == "helm-oci") | .valuesFiles // [] | map(. as $item | ($item.version | split(".")) as $version | {major: $version[0], minor: $version[1], patch: $version[2], version: $item.version} ) | sort_by(.major, .minor, .patch) | last | .version // "0.0.0"')
    fi

    keys=$(yq -o json $input | jq -rc --arg repository $repository '.[] | select(.repository == $repository and .kind == "helm-oci") | .valuesFiles // [] | .[].version // "0.0.0"')

    if echo "$keys" | grep "$version" >/dev/null; then
        relevant_version=$version
    else
        relevant_version=$({ echo $version; echo $keys; } | tr ' ' "\n" | sort -t "." -k1,1n -k2,2n -k3,3n | grep -B1 $version | grep -v $version)
    fi

    yq -o json $input | jq -rcj --arg repository $repository --arg version "$relevant_version" '.[] | select(.repository == $repository and .kind == "helm-oci") | .valuesFiles // [] | .[] | select(.version == $version) | .valuesFile // ""'
}
