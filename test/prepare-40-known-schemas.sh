echo "Setup known schemas ..."

if [ "$1" = "only-latest" ]; then
    mkdir -p /schema/chart.local/ &>/dev/null || true
    mkdir -p /schema/chart.new/ &>/dev/null || true
    mkdir -p /schema/chart.uri/ &>/dev/null || true
    mkdir -p /schema/chart.unconditional/ &>/dev/null || true
    mkdir -p /schema/chart.local-oci/ &>/dev/null || true
    mkdir -p /schema/chart.new-oci/ &>/dev/null || true
    mkdir -p /schema/chart.unconditional-oci/ &>/dev/null || true
    mkdir -p /schema/source.example.com/ &>/dev/null || true
fi

echo "  - updated for test '$1'"
echo
