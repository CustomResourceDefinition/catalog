set -e
echo "Sorting contents of schema validations ..."
for directory in "$2"/*/; do
    echo "  - $(basename $directory)"
    cd $directory
    find . -name "*.json" -type f -print0 | sort -z | xargs -0 -I{} sh -c 'jq -S < "{}" > "{}.tmp"; mv "{}.tmp" "{}"'
    cd - >/dev/null
done
echo
