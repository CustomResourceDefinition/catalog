set -e
echo "Converting files to schema validation ..."
for directory in "$2"/*/; do
    echo "  - $(basename $directory)"
    cd $directory
    find . -name "*.yaml" -type f -print0 | sort -z | xargs -0 -I{} sh -c 'python /app/helpers/convert.py "{}" >/dev/null; rm "{}"'
    cd - >/dev/null
done

echo "Cleaning up problematic files:"
find "$2" -name "*.yaml" -print0 | xargs -0 -I{} sh -c 'rm "{}"; echo "  - {}"'
echo "  - done"
echo
