# CRD Catalog

This repository aggregates hundreds of popular Kubernetes CRDs (`CustomResourceDefinition`) in JSON schema format. These schemas can be used by various tools such as [Kubeconform](https://github.com/yannh/kubeconform) as an alternative to `kubectl --dry-run` to perform validation on custom (and native) Kubernetes resources.

Running Kubernetes schema validation checks helps apply the **"shift-left approach"** on machines **without** giving them access to your cluster (e.g. locally or on CI).

## Origins

This catalog is inspired by the [CRDs-catalog](https://github.com/datreeio/CRDs-catalog) from [Datree](https://www.datree.io), but instead uses a sources list to pull updates automatically and enables re-creation from these sources provided the source helm charts, uris, etc are still available.

## How to use the schemas in the catalog

### Kubeconform
```sh
kubeconform -schema-location default -schema-location 'https://raw.githubusercontent.com/CustomResourceDefinition/catalog/main/schema/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' [MANIFEST]
```

# How to contribute CRDs

You need to create a pull request with changes to the sources list using the methods below:

* [Helm charts](#helm-charts) - the preferred method
* [Git](#git)
* [OCI charts](#oci-charts)
* [Uris](#uris) - when everything else fails

We prefer using the [Helm charts](#helm-charts) method to avoid issues like needing to specify each release version with CRD changes or version history being unavailable.

> [!IMPORTANT]  
> Please keep the `configuration.yaml` sorted by name.  

## Helm charts

To add CRDs for [ArgoCD](https://github.com/argoproj/argo-cd) you should apply the following changes to `configuration.yaml`.

```yaml
- entries:
    - argo-cd
  kind: helm
  name: argo
  repository: https://argoproj.github.io/argo-helm
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
  versionPrefix: v # by default only major.minor.patch tags are used, but a prefix can be set
  # versionSuffix: # regex matching end of versions, by default matching $
  # includeHead: true # by default the head branch is ignored and only published tags are used
  # searchPaths: # paths to recursively find yaml files in (non-CRDs are discarded)
  #   - crds
  # genPaths: # paths to recursively find go files to generate CRDs from
  #   - api/...
```

## OCI charts

> [!IMPORTANT]  
> Always use a https helm repository, if one is available.  

To add CRDs for [CrunchyData/postgres-operator](https://github.com/CrunchyData/postgres-operator) you should apply the following changes to `configuration.yaml`.

```yaml
- additionalVersions:
    - 5.5.2
  kind: helm-oci
  name: crunchydata-pgo
  repository: oci://registry.developers.crunchydata.com/crunchydata/pgo
```

The `additionalVersions` entry should be a list of previous versions that have old CRDs that should still be available. You should always add the current version to this list when initially adding a new chart.

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
