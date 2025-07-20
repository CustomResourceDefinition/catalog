echo "Setup known schemas ..."

if [ "$1" = "only-latest" ]; then
    cd /verified-schema
    find . -print | while read -r item; do
        if [ -d "$item" ]; then
            mkdir -p "/schema/$item"
        elif [ -f "$item" ]; then
            touch "/schema/$item"
        fi
    done
    cd -
fi

echo "  - updated for test '$1'"
echo
