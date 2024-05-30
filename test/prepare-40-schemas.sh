printf "Setup known schemas ... "

if [ "$1" = "only-latest" ]; then
    mkdir -p /schema/chart.local/ &>/dev/null || true
    mkdir -p /schema/chart.new/ &>/dev/null || true
    mkdir -p /schema/chart.uri/ &>/dev/null || true
    echo "Updated for test '$1'"
fi

echo
