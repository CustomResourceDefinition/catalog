set -e

printf "Reset for a fresh test ... "

rm -rf /schema/* &>/dev/null || true
rm -rf /root/.cache/helm/* &>/dev/null || true
rm -rf /root/.config/helm/* &>/dev/null || true
rm -rf /root/.local/helm/* &>/dev/null || true
rm -rf /templates/* &>/dev/null || true
rm -rf /repository &>/dev/null || true
rm -rf /verified-schema &>/dev/null || true

mkdir /verified-schema
mkdir /repository
mkdir /repository/git
mkdir /repository/http

cd /repository/http
killall -9 python3 &> /dev/null || true
(python3 -m http.server 2>&1 | grep -v '" 200 -' | grep -v 'Connection refused') &

echo "done"
echo
