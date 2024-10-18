set -e

echo "Setup verified schemas ... "

rm -rf -- /tmp/verified &>/dev/null || true
mkdir -p /tmp/verified/base
cp /app/test/fixtures/*.json /tmp/verified/base/

if [ "$1" = "all-versions" ]; then
    cp -r /tmp/verified/base /verified-schema/chart.conditional
    cp -r /tmp/verified/base /verified-schema/chart.unconditional
    cp -r /tmp/verified/base /verified-schema/chart.old
    cp -r /tmp/verified/base /verified-schema/chart.local
    cp -r /tmp/verified/base /verified-schema/chart.conditional-oci
    cp -r /tmp/verified/base /verified-schema/chart.unconditional-oci
    cp -r /tmp/verified/base /verified-schema/chart.old-oci
    cp -r /tmp/verified/base /verified-schema/chart.local-oci
    cp -r /tmp/verified/base /verified-schema/chart.git
    cp -r /tmp/verified/base /verified-schema/chart.uri
    cp -r /tmp/verified/base /verified-schema/chart.uri1
elif [ "$1" = "only-latest" ]; then
    cp -r /tmp/verified/base /verified-schema/chart.unconditional
    cp -r /tmp/verified/base /verified-schema/chart.local
    cp -r /tmp/verified/base /verified-schema/chart.unconditional-oci
    cp -r /tmp/verified/base /verified-schema/chart.local-oci
    cp -r /tmp/verified/base /verified-schema/chart.git
    cp -r /tmp/verified/base /verified-schema/chart.uri
fi

echo "  - updated for test '$1'"
echo
