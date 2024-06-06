
tmp=$(dirname "$1")/$(basename -s .yaml "$1").tmp
template=$(dirname "$1")/$(basename -s .yaml "$1")-part

mv "$1" $tmp
python3 $(dirname "$0")/split.py "$tmp" "$template"
rm "$tmp"
