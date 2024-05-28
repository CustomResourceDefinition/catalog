# FIXME: break on errors?
echo "Arranging manifests ..."
for directory in /templates/*/; do
    echo "  - $(basename $directory)"
    find $directory -name "*.yaml" -type f -print0 | sort -z | while IFS= read -r -d '' file; do
        group=$(yq .spec.group < $file)
        if [ $group = "null" ]; then
            rm $file
            continue
        fi
        mkdir -p "/schema/$group" || true
        mv $file "/schema/$group/"
    done
done
