# CRD Catalog

This repository aggregates hundreds of popular Kubernetes CRDs (`CustomResourceDefinition`) in JSON schema format. These schemas can be used by various tools such as [Kubeconform](https://github.com/yannh/kubeconform) as an alternative to `kubectl --dry-run` to perform validation on custom (and native) Kubernetes resources.

Running Kubernetes schema validation checks helps apply the **"shift-left approach"** on machines **without** giving them access to your cluster (e.g. locally or on CI).

## Origins

This catalog is inspired by [Datrees CRDs-catalog](https://github.com/datreeio/CRDs-catalog), but uses sources list to pull updates automatically and enables re-creation from these source provided the source helm charts, uris, etc are still available.

## How to use the schemas in the catalog

### Kubeconform
```sh
kubeconform -schema-location default -schema-location 'https://raw.githubusercontent.com/CodeReaper/CRD-catalog/main/schema/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' [MANIFEST]
```

# How to contribute CRDs

You need to create a pull request with changes to the sources list(s) using the methods below:

* [Helm charts](#helm-charts) - the preferred method
* [Uris](#uris)

We prefer using the [Helm charts](#helm-charts) method to avoid needing to specify each release version with CRD changes which is required by the [Uris](#uris) method.

## Helm charts

To add CRDs for [ArgoCD](https://github.com/argoproj/argo-cd) you should apply the following changes to `helm-charts.yaml`.

```yaml
- repository: https://argoproj.github.io/argo-helm
  name: argo
  entries:
    - argo-cd
```

> [!NOTE]  
> Each `repository` should only be listed once in the file.  
> The `name` of a repository should be the name given by the developers (if possible).  

> [!IMPORTANT]  
> Always add the repository that the developers publish to directly and avoid aggregated helm chart repositories like Truechart etc.  

If the above example already exists when you are trying to add [Argo Rollouts](https://github.com/argoproj/argo-rollouts). You should only add another chart entry to the existing repository entry, so the repository entry looks like the following.

```yaml
- repository: https://argoproj.github.io/argo-helm
  name: argo
  entries:
    - argo-cd
    - argo-rollouts
```

## Uris

To add CRDs for version 1.0.0 of the fictional Custom tool you should apply the following changes to `manifest-uris.yaml`.

```yaml
- name: custom
  apiGroups:
    - custom.io
  crds:
    - version: 1.0.0
      baseUri: https://raw.githubusercontent.com/crds-r-us/custom/v1.0.0/chart/template/crds
      paths:
        - custom-crd.yaml
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
