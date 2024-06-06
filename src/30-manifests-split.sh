echo "Spliting all collected CRDs into separate files ..."
find "$3/templates" -name "*.yaml" -type f -exec sh build/bin/helpers/split.sh {} \;
find "$3/templates" -name "*.yaml" -type f -empty -delete
echo
