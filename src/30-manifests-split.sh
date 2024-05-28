echo "Split them files"
exit 0

output=crds
mkdir -p $output
python /app/helpers/split.py crds.yaml "${output}/f"
