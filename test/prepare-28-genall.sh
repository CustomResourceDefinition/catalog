set -e

echo "Setup test genall source ... "

mkdir -p /verified-schema/source.example.com
cp /app/test/fixtures/foo_foo.json /verified-schema/source.example.com/
cp /app/test/fixtures/foo.yaml /verified-schema/source.example.com/foo.yaml

echo
