echo "Setup known schemas ..."

if [ "$1" = "only-latest" ]; then
    rm -r /verified-schema/chart.conditional
    rm -r /verified-schema/chart.conditional-oci
    rm -r /verified-schema/chart.old
    rm -r /verified-schema/chart.old-oci
    rm -r /verified-schema/chart.uri1

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
