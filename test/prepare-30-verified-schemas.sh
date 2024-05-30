set -e

printf "Setup verified schemas ... "

rm -rf /verified-schema &>/dev/null || true
cp -r /verified-schemas/$1 /verified-schema
echo "$1"
echo


