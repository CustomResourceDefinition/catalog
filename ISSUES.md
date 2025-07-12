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

## Clean configuration

- remove all '...' from genpaths