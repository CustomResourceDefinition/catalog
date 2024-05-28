echo "Convert CRDs to schema validation json files ..."
find /templates -name "*.yaml" -type f -exec sh -c 'python /app/helpers/convert.py $0; rm $0' {} \;
