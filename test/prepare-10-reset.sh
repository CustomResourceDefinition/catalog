set -e

echo "Reset for a fresh test ... "

helm repo ls -o yaml | yq -r '.[].name' | xargs -I {} helm repo rm {}

rm -rf /schema/* &>/dev/null || true
rm -rf /templates/* &>/dev/null || true
rm -rf /repository &>/dev/null || true
rm -rf /verified-schema &>/dev/null || true
rm -rf /tmp/verified &>/dev/null || true

mkdir -p /verified-schema
mkdir -p /repository/git
mkdir -p /tmp/verified/base

echo "done"
echo
