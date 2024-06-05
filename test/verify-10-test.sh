set -e

echo "Verifing $1 ..."

find "$2/schema" -type f -name "*.json" -exec sh -c 'echo diff $0 ${0/schema/verified-schema}' {} \; | sh -e
find "$3/verified-schema" -type f -name "*.json" -exec sh -c 'echo diff $0 ${0/verified-schema/schema}' {} \; | sh -e

echo
echo " --- Passed - $1"
echo
