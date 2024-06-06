set -e

echo "Verifing $1 ..."

find "$3/schema" -type f -name "*.json" -exec sh -c 'echo diff $0 $(echo "$0" | sed s/schema/verified-schema/)' {} \; | sh -e
find "$3/verified-schema" -type f -name "*.json" -exec sh -c 'echo diff $0 $(echo "$0" | sed s/verified-schema/verified-schema/)' {} \; | sh -e

echo
echo " --- Passed - $1"
echo
