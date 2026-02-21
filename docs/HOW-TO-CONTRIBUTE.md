# How to contribute CRDs

You need to create a pull request with changes to the sources list using the methods (prepared in order of preference) below:

- [Helm charts](#helm-charts)
- [OCI charts](#oci-charts)
- [Git](#git)
- [Uris](#uris) - when everything else fails

We prefer using the [Helm charts](#helm-charts) method to avoid issues like needing to specify each release version with CRD changes or version history being unavailable.

> [!IMPORTANT]  
> Please keep the `configuration.yaml` sorted by name.

We encourage you to look through the [COMPARISON.md](COMPARISON.md), if you would like to help add relevant missing schemas.

## Helm charts

To add CRDs for [ArgoCD](https://github.com/argoproj/argo-cd) you should apply the following changes to `configuration.yaml`.

```yaml
- entries:
    - argo-cd
  kind: helm
  name: argo
  repository: https://argoproj.github.io/argo-helm
  # versionPattern: # regex pattern with capture group for semver major.minor.patch (default: ^v?([0-9]+\.[0-9]+\.[0-9]+)$)
```

> [!NOTE]  
> Each `repository` should only be listed once in the file, if possible.  
> The `name` of a repository should be the name given by the developers, if possible.

> [!IMPORTANT]  
> Always add the repository that the developers publish to directly and avoid aggregated helm chart repositories like Truechart etc.

If the above example already exists when you are trying to add [Argo Rollouts](https://github.com/argoproj/argo-rollouts). You should only add another chart entry to the existing repository entry, so the repository entry looks like the following.

```yaml
- entries:
    - argo-cd
    - argo-rollouts
  kind: helm
  name: argo
  repository: https://argoproj.github.io/argo-helm
```

## Git

To add CRDs for [eks-anywhere](https://github.com/aws/eks-anywhere) you should apply the following changes to `configuration.yaml`.

```yaml
- kind: git
  kustomizationPaths: # paths to build kustomizations from (non-CRDs are discarded)
    - config/crd
  name: eks-anywhere
  repository: https://github.com/aws/eks-anywhere
  # versionPattern: # regex pattern with capture group for semver major.minor.patch (default: ^([0-9]+\.[0-9]+\.[0-9]+)$)
  # searchPaths: # paths to recursively find yaml files in (non-CRDs are discarded)
  #   - crds
  # genPaths: # paths to recursively find go files to generate CRDs from
  #   - api
```

The `versionPattern` is a regex that must contain a capture group for the semver major.minor.patch version.

Examples:

- Match v-prefixed tags: `^v([0-9]+\.[0-9]+\.[0-9]+)$`
- Match tags with custom prefix: `^prefix-([0-9]+\.[0-9]+\.[0-9]+)$`
- Match tags with pre-release suffixes: `^v([0-9]+\.[0-9]+\.[0-9]+-.*)$`
- Match versioned tags and branches: `^(([0-9]+\.[0-9]+\.[0-9]+)|(main|master))$`
- Match branches only: `^((main|master))$`

Git sources are sorted by commit date (most recent first), not by semver.

## OCI charts

To add CRDs for [CrunchyData/postgres-operator](https://github.com/CrunchyData/postgres-operator) you should apply the following changes to `configuration.yaml`.

```yaml
- kind: helm-oci
  name: crunchydata-pgo
  repository: oci://registry.developers.crunchydata.com/crunchydata/pgo
  # versionPattern: # regex pattern with capture group for semver major.minor.patch (default: ^v?([0-9]+\.[0-9]+\.[0-9]+)$)
```

## Uris

To add CRDs for version 1.0.0 of the fictional Custom tool you should apply the following changes to `configuration.yaml`.

```yaml
- apiGroups:
    - custom.io
  crds:
    - baseUri: https://raw.githubusercontent.com/crds-r-us/custom/v1.0.0/chart/template/crds
      paths:
        - custom-crd.yaml
      version: 1.0.0
  kind: http
  name: custom
```

> [!NOTE]  
> Each `name` should only be listed once in the file.

### `apiGroups`

This entry should contain a complete list of unique value specified in the CRDs under the yaml path `.spec.group`.

> [!IMPORTANT]  
> You should only ever append to this list.

### `crds`

Each release version that contains addition, changes or removal to CRDs must be listed separately in this list.

The `baseUri` combined with each entry in `paths` using the template `${baseUri}/${pathEntry}` must be a URI to a manifest file that is publicly available. You can split the uris as you want, for instance to optimize for the longest common `baseUri`.

> [!TIP]
> Each release version has their own `baseUri` and `paths` entries.
