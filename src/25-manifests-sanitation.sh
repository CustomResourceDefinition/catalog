tmp=$(mktemp -d)
base=$(pwd)
set -e
cd "$3/templates"

echo "Sanitizing manifests ..."

echo "  - removing no content manifests:"
find . -type f -exec sh "$base/build/bin/helpers/delete-whitespace-file.sh" {} \;
echo "    - done"

echo "  - removing descriptions from manifests:"
for directory in */*/; do
    echo "    - $directory"
    find $directory -name "*.yaml" -type f -print0 | sort -z | xargs -0 -I{} sh "$base/build/bin/helpers/remove-description.sh" "{}"
done
echo "    - done"

echo "  - deduplicating:"
for directory in */*/; do
    find $directory -type f -exec md5sum {} + > $tmp/file_hashes.txt
    sort $tmp/file_hashes.txt > $tmp/sorted_hashes.txt
    print=1
    for file in $(uniq -w32 -d $tmp/sorted_hashes.txt | cut -d' ' -f3-); do
        rm "$file"
        if [ $print -eq 1 ]; then
            print=0
            echo "    - $directory"
        fi
    done
done
echo "    - done"

cd - >/dev/null
echo
