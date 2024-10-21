set -e

echo "Reset for a fresh test ... "

helm repo ls -o json | jq -r '.[].name' | xargs -I {} helm repo rm {}

rm -rf /schema/* &>/dev/null || true
rm -rf /templates/* &>/dev/null || true
rm -rf /repository &>/dev/null || true
rm -rf /verified-schema &>/dev/null || true

mkdir /verified-schema
mkdir /repository/git

echo "done"
echo
