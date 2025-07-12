# Problems

## helm `--include-crds`?
```
- entries:
    - cert-manager
  kind: helm
  name: cert-manager
  repository: https://charts.jetstack.io
  valuesFiles:
    - version: 0.0.0
      valuesFile: |
        installCRDs: true
    - version: 1.15.0
      valuesFile: |
        crds:
          enabled: true
```

## oci
```
- additionalVersions:
    - 5.5.2
  kind: helm-oci
  name: crunchydata-pgo
  repository: oci://registry.developers.crunchydata.com/crunchydata/pgo
- additionalVersions:
    - 1.0.6
  kind: helm-oci
  name: karpenter
  repository: oci://public.ecr.aws/karpenter/karpenter
```
