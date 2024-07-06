set -e

printf '- entries:
    - templated
  kind: helm
  name: versioned
  repository: repository
  valuesFiles:
    - version: 0.0.0
      valuesFile: |
        output: true
    - version: 2.0.0
      valuesFile: |
    - version: 1.0.0
      valuesFile: |
        output: true
- entries:
    - regular
  kind: helm
  name: unversioned
  repository: repository
' > /tmp/input

echo 'output: true' > /tmp/expected
values_file_of /tmp/input repository versioned 0.0.0 > /tmp/output
diff /tmp/expected /tmp/output

echo 'output: true' > /tmp/expected
values_file_of /tmp/input repository versioned 1.0.0 > /tmp/output
diff /tmp/expected /tmp/output

echo 'output: true' > /tmp/expected
values_file_of /tmp/input repository versioned 1.999.999 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of /tmp/input repository versioned 2.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of /tmp/input repository versioned 2.999.999 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of /tmp/input repository versioned > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of /tmp/input repository unversioned 0.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of /tmp/input repository unversioned 1.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of /tmp/input repository unversioned 2.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of /tmp/input repository unversioned > /tmp/output
diff /tmp/expected /tmp/output
