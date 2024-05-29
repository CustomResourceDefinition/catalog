# FIXME: break on errors?
echo "Sorting contents of schema validations ..."
for directory in /schema/*/; do
    echo "  - $(basename $directory)"
    cd $directory
    find . -name "*.json" -type f -print0 | sort -z | xargs -0 -I{} sh -c 'jq -S < "{}" > "{}.tmp"; mv "{}.tmp" "{}"'
    cd - >/dev/null
done
