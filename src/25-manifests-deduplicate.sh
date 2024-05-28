tmp=$(mktemp -d)
# FIXME: break on errors?
echo "Deduplicating ..."
for directory in /templates/*/; do
    name=$(basename $directory)
    echo "  - $name"

    find $directory -type f -exec md5sum {} + > $tmp/file_hashes.txt
    sort $tmp/file_hashes.txt > $tmp/sorted_hashes.txt
    for file in $(uniq -w32 -d $tmp/sorted_hashes.txt | cut -d' ' -f3-); do
        rm "$file"
    done
done
