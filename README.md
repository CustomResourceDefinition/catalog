# CRD Catalog

This repository aggregates hundreds of popular Kubernetes CRDs (`CustomResourceDefinition`) in JSON schema format. These schemas can be used by various tools such as [Kubeconform](https://github.com/yannh/kubeconform) as an alternative to `kubectl --dry-run` to perform validation on custom (and native) Kubernetes resources.

Running Kubernetes schema validation checks helps apply the **"shift-left approach"** on machines **without** giving them access to your cluster (e.g. locally or on CI).

## How to use the schemas in the catalog

### Kubeconform
```sh
kubeconform -schema-location default -schema-location 'https://raw.githubusercontent.com/CodeReaper/CRD-catalog/main/schema/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' [MANIFEST]
```

## Contributing CRDs to the catalog
If the catalog is missing public custom resources (CRDs) that you would like to automatically validate using these tools, you can open an issue or create a pull request to add the helm chart to this repository.
