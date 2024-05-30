# FIXME: break on errors?
echo "Arranging manifests ..."
cd /templates
for directory in */*/; do
    echo "  - $directory"
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
cd - >/dev/null
echo
