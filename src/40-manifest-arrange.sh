set -e
echo "Arranging manifests ..."

for directory in "$3/templates"/*/*/; do
    echo "  - $(echo "$directory" | rev | cut -d/ -f1-3 | rev)"
    find $directory -name "*.yaml" -type f -print0 | sort -z | xargs -0 -I{} sh build/bin/helpers/arrange.sh "{}" "$2"
done
echo
