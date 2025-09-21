set -e

echo "Verifing $1 ..."

{
    cd /schema || true
    find . -type f \( -name "*.json" -o -name "*.yaml" \)
    cd - &>/dev/null || true
} | sort > /tmp/schema.list

{
    cd /verified-schema || true
    find . -type f \( -name "*.json" -o -name "*.yaml" \)
    cd - &>/dev/null || true
} | sort > /tmp/verified.list

diff /tmp/schema.list /tmp/verified.list

for f in $(find /verified-schema -type f \( -name "*.json" -o -name "*.yaml" \)); do
    diff "/${f#/verified-}" "$f"
done

echo
echo " --- Passed - $1"
echo
