set -e

printf "Reset for a fresh test ... "

rm -rf "$2"/* &>/dev/null || true
# rm -rf /root/.cache/helm/* &>/dev/null || true
# rm -rf /root/.config/helm/* &>/dev/null || true
# rm -rf /root/.local/helm/* &>/dev/null || true
rm -rf "$3/templates"/* &>/dev/null || true
rm -rf "$3/repository" &>/dev/null || true
rm -rf "$3/verified-schema" &>/dev/null || true

mkdir "$3/verified-schema"
mkdir "$3/repository"
mkdir "$3/repository/git"
mkdir "$3/repository/http"

cd "$3/repository/http"
killall -9 python3 &> /dev/null || true
(python3 -m http.server 2>&1 | grep -v '" 200 -' | grep -v 'Connection refused') &
cd - &>/dev/null

echo "done"
echo
