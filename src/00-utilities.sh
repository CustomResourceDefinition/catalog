values_file_of() {
    local input=$1
    local repository=$2
    local name=$3
    local version=$4
    local relevant_version

    if [ "" = "$4" ]; then
        version=$(yq -o json $input | jq -rc --arg repository $repository --arg name $name '.[] | select(.repository == $repository and .name == $name) | .valuesFiles // [] | map(. as $item | ($item.version | split(".")) as $version | {major: $version[0], minor: $version[1], patch: $version[2], version: $item.version} ) | sort_by(.major, .minor, .patch) | last | .version // "0.0.0"')
    fi

    keys=$(yq -o json $input | jq -rc --arg repository $repository --arg name $name '.[] | select(.repository == $repository and .name == $name) | .valuesFiles // [] | .[].version // "0.0.0"')

    if echo "$keys" | grep "$version" >/dev/null; then
        relevant_version=$version
    else
        relevant_version=$({ echo $version; echo $keys; } | tr ' ' "\n" | sort -t "." -k1,1n -k2,2n -k3,3n | grep -B1 $version | grep -v $version)
    fi

    yq -o json $input | jq -rcj --arg repository $repository --arg name $name --arg version $relevant_version '.[] | select(.repository == $repository and .name == $name) | .valuesFiles // [] | .[] | select(.version == $version) | .valuesFile // ""'
}
