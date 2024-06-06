echo "Setup known schemas ..."

if [ "$1" = "only-latest" ]; then
    mkdir -p "$2/chart.local/" > /dev/null 2>&1 || true
    mkdir -p "$2/chart.old/" > /dev/null 2>&1 || true
    mkdir -p "$2/chart.uri/" > /dev/null 2>&1 || true
fi

echo "  - updated for test '$1'"
echo
