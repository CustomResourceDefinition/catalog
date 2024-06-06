if grep -q "[^[[:space:]]]" "$1"; then
    rm "$1"
    echo "    - $1"
fi
