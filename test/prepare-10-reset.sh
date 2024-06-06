set -e

echo "Reset for a fresh test ... "

helm repo ls -o json | jq -r '.[].name' | xargs -I {} helm repo rm {}

rm -rf /schema/* &>/dev/null || true
rm -rf /templates/* &>/dev/null || true
rm -rf /repository &>/dev/null || true
rm -rf /verified-schema &>/dev/null || true

mkdir /verified-schema
mkdir /repository
mkdir /repository/git
mkdir /repository/http

cd /repository/http
pgrep -f python3 | xargs -r kill -9 &> /dev/null || true
(python3 -m http.server 2>&1 | grep -v '" 200 -' | grep -v 'Connection refused') &

echo "done"
echo
