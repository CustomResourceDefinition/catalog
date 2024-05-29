set -e

printf "Setup test schema ... "

rm -rf /verified-schema &>/dev/null || true
cp -r /verified-schemas/$1 /verified-schema
echo "$1"
echo
