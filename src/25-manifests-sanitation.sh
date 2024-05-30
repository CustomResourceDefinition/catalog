tmp=$(mktemp -d)
set -e
cd /templates

echo "Sanitizing manifests ..."

echo "  - removing no content manifests:"
for directory in */*/; do
    #shellcheck disable=SC2038
    find $directory -type f -exec sh -c 'grep -q "[^[:space:]]" "$0" || echo "$0"' {} \; | xargs -I{} sh -c 'rm "{}"; echo "    - {}"'
    #shellcheck enable=SC2038
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
