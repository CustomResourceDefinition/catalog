group=$(yq .spec.group < $1)
if [ $group = "null" ]; then
    rm $1
    exit 0
fi
mkdir -p "$2/$group" || true
mv $1 "$2/$group/"
