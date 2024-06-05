echo "Spliting all collected CRDs into separate files ..."
find "$3/templates" -name "*.yaml" -type f -exec sh -c 'mv $0 ${0/.yaml/.tmp}' {} \;
find "$3/templates" -name "*.tmp" -type f -exec sh -c 'python build/bin/helpers/split.py $0 ${0/.tmp/-part}; rm $0' {} \;
find "$3/templates" -name "*.yaml" -type f -empty -delete
echo
