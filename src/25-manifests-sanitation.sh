tmp=$(mktemp -d)
# FIXME: break on errors?
echo "Sanitizing manifests ..."

echo "  - deduplicating:"
cd /templates
for directory in */*/; do
    echo "    - $directory"

    find $directory -type f -exec md5sum {} + > $tmp/file_hashes.txt
    sort $tmp/file_hashes.txt > $tmp/sorted_hashes.txt
    for file in $(uniq -w32 -d $tmp/sorted_hashes.txt | cut -d' ' -f3-); do
        rm "$file"
    done
done

for directory in */*/; do
    name=$(basename $directory)
    find $directory -type f -empty -delete >/dev/null
done

echo "  - no output generated:"
for directory in */*/; do
    find $directory -type f -print -quit | grep -q . || echo "    - $directory"
done

cd - >/dev/null
echo
