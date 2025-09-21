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

for f in $(find /verified-schema -type f -name "*.json"); do
    diff "/${f#/verified-}" "$f"
done

set +e
for f in $(find /verified-schema -type f -name "*.yaml"); do
    if ! diff <(yq -o=json "/${f#/verified-}" | jq -S | yq -P) <(yq -o=json "$f" | jq -S | yq -P); then # complex diff for comparison regardless of whitespace and ordering
        echo
        echo
        echo "Failed check on $f"
        exit 1
    fi
done

echo
echo " --- Passed - $1"
echo
