
echo "Verifing test for '$1' ..."
set -e
find /schema -type f -name "*.json" -exec sh -c 'echo diff $0 ${0/schema/verified-schema}' {} \; | sh -e
find /verified-schema -type f -name "*.json" -exec sh -c 'echo diff $0 ${0/verified-schema/schema}' {} \; | sh -e

echo "Tests passed - $1"
