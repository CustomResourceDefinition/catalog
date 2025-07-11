set -e

echo "Verifing $1 ..."

{
  cd /schema || true
  find . -type f -name "*.json"
  cd - &>/dev/null || true
} | sort > /tmp/schema.list

{
  cd /verified-schema || true
  find . -type f -name "*.json"
  cd - &>/dev/null || true
} | sort > /tmp/verified.list

diff /tmp/schema.list /tmp/verified.list

echo
echo " --- Passed - $1"
echo
