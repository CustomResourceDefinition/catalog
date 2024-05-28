echo "Spliting all collected CRDs into separate files ..."
find /templates -name "*.yaml" -type f -exec sh -c 'mv $0 ${0/.yaml/.tmp}' {} \;
find /templates -name "*.tmp" -type f -exec sh -c 'python /app/helpers/split.py $0 ${0/.tmp/-part}; rm $0' {} \;
find /templates -name "*.yaml" -type f -empty -delete
