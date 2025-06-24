set -e
echo "Converting files to schema validation ..."
for directory in /schema/*/; do
    echo "  - $(basename $directory)"
    cd $directory
    find . -name "*.yaml" -type f -print0 | sort -z | xargs -0 -I{} sh -c 'catalog convert --input "{}" --output ./; rm "{}"'
    cd - >/dev/null
done

echo "Cleaning up problematic files:"
find /schema -name "*.yaml" -print0 | xargs -0 -I{} sh -c 'rm "{}"; echo "  - {}"'
echo "  - done"
echo
