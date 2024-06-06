set -e
echo "Arranging manifests ..."
for directory in "$3/templates"/*/*/; do
    echo "  - $(echo "$directory" | rev | cut -d/ -f1-3 | rev)"
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
echo
