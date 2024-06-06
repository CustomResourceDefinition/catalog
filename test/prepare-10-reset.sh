set -e

echo "Reset for a fresh test ... "

helm repo ls -o json | jq -r '.[].name' | xargs -I{} helm repo rm {}

rm -rf "${3:?}"/* > /dev/null 2>&1 || true

mkdir "$3/schema"
mkdir "$3/verified-schema"
mkdir "$3/repository"
mkdir "$3/repository/git"
mkdir "$3/repository/http"

cd "$3/repository/http"
ps -ef | grep python3 | grep -v grep | tr -s ' ' | cut -d ' ' -f 2 | xargs -r kill -9 > /dev/null 2>&1 || true
(python3 -m http.server 2>&1 | grep -v '" 200 -' | grep -v 'Connection refused' | grep -v 'Killed') &
cd - > /dev/null 2>&1

echo "done"
echo
