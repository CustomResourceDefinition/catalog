# Problems

## oci
```yaml
- additionalVersions:
    - 5.5.2
  kind: helm-oci
  name: crunchydata-pgo
  repository: oci://registry.developers.crunchydata.com/crunchydata/pgo # not functional but for current hacky code, but works for ghcr.io, public.ecr.aws
- additionalVersions:
    - 1.0.6
  kind: helm-oci
  name: karpenter
  repository: oci://public.ecr.aws/karpenter/karpenter-crd # this is a new repo - should be be run without `schema/karpenter.sh`
```

## smoke test #2

Too much is produced:

```
Verifing Works using only latest version ...
--- /tmp/schema.list
+++ /tmp/verified.list
@@ -1,12 +1,7 @@
-./chart.conditional-oci/test_v1.json
-./chart.conditional/test_v1.json
 ./chart.git/test_v1.json
 ./chart.local-oci/test_v1.json
 ./chart.local/test_v1.json
-./chart.old-oci/test_v1.json
-./chart.old/test_v1.json
 ./chart.unconditional-oci/test_v1.json
 ./chart.unconditional/test_v1.json
 ./chart.uri/test_v1.json
-./chart.uri1/test_v1.json
 ./source.example.com/foo_foo.json
make[1]: *** [make.d/tests.mk:55: _smoke-test] Error 1
```