set -e

echo 'output: true' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository versioned 0.0.0 > /tmp/output
diff /tmp/expected /tmp/output

echo 'output: true' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository versioned 1.0.0 > /tmp/output
diff /tmp/expected /tmp/output

echo 'output: true' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository versioned 1.999.999 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository versioned 2.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository versioned 2.999.999 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository versioned > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository unversioned 0.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository unversioned 1.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository unversioned 2.0.0 > /tmp/output
diff /tmp/expected /tmp/output

printf '' > /tmp/expected
values_file_of test/fixtures/values-files.yaml repository unversioned > /tmp/output
diff /tmp/expected /tmp/output
