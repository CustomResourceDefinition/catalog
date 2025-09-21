set -e

echo "Setup verified schemas ... "

cp /app/test/fixtures/test_v1.json /tmp/verified/base/
cp /app/test/fixtures/test-crd.yaml /tmp/verified/base/test.yaml

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

mkdir -p /verified-schema/source.example.com
cp /app/test/fixtures/foo_foo.json /verified-schema/source.example.com/
touch /verified-schema/source.example.com/foo.yaml # FIXME: argh

echo "  - updated for test '$1'"
echo
