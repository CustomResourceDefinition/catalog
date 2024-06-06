set -e
echo "Sorting contents of schema validations ..."
for directory in "$2"/*/; do
    echo "  - $(basename $directory)"
    find "$directory" -name "*.json" -type f -print0 | sort -z | xargs -0 -I{} sh -c 'jq -S < "{}" > "{}.tmp"; mv "{}.tmp" "{}"'
done
echo "  - done"
echo
