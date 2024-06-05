set -e
echo "Arranging manifests ..."
cd "$3/templates"
for directory in */*/; do
    echo "  - $directory"
    find $directory -name "*.yaml" -type f -print0 | sort -z | while IFS= read -r -d '' file; do
        group=$(yq .spec.group < $file)
        if [ $group = "null" ]; then
            rm $file
            continue
        fi
        mkdir -p "$2/$group" || true
        mv $file "$2/$group/"
    done
done
cd - >/dev/null
echo
