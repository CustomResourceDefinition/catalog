set -e

echo "Setup verified schemas ... "

rm -rf -- /tmp/verified > /dev/null 2>&1 || true
mkdir -p /tmp/verified/base
cp test/fixtures/*.json /tmp/verified/base/

if [ "$1" = "all-versions" ]; then
    cp -r /tmp/verified/base "$3/verified-schema/chart.conditional"
    cp -r /tmp/verified/base "$3/verified-schema/chart.git"
    cp -r /tmp/verified/base "$3/verified-schema/chart.old"
    cp -r /tmp/verified/base "$3/verified-schema/chart.local"
    cp -r /tmp/verified/base "$3/verified-schema/chart.uri"
    cp -r /tmp/verified/base "$3/verified-schema/chart.uri1"
elif [ "$1" = "only-latest" ]; then
    cp -r /tmp/verified/base "$3/verified-schema/chart.conditional"
    cp -r /tmp/verified/base "$3/verified-schema/chart.git"
    cp -r /tmp/verified/base "$3/verified-schema/chart.local"
    cp -r /tmp/verified/base "$3/verified-schema/chart.uri"
fi

echo "  - updated for test '$1'"
echo
