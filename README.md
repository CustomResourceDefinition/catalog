# CRD Catalog

This repository aggregates hundreds of popular Kubernetes CRDs (`CustomResourceDefinition`) including their JSON schema format.

The intended purpose with this repository is aid with validation and code generation.

The catalog are checked for updates every 8 hours, but the catalog is also tagged with kubernetes versions to make the catalog available as it were when a specific version was released.

## Known use cases

### Validation using Kubeconform

The schema files can be used by various tools such as [Kubeconform](https://github.com/yannh/kubeconform) as an alternative to `kubectl --dry-run` to perform validation on custom (and native) Kubernetes resources.

Running Kubernetes schema validation checks helps apply the **"shift-left approach"** on machines **without** giving them access to your cluster (e.g. locally or on CI).

Example:

```sh
kubeconform -schema-location default -schema-location 'https://raw.githubusercontent.com/CustomResourceDefinition/catalog/main/schema/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' [MANIFEST]
```

Example using the catalog as it were when version 1.33.6 was released:

```sh
kubeconform -schema-location default -schema-location 'https://raw.githubusercontent.com/CustomResourceDefinition/catalog/refs/tags/v1.33.6/schema/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' [MANIFEST]
```

### Code generation using kopium

The definition files can be used by tools like [kopium](https://github.com/kube-rs/kopium) to generate Rust data structs.

Example:

```sh
curl -sSL 'https://raw.githubusercontent.com/CustomResourceDefinition/catalog/main/definitions/monitoring.coreos.com/prometheusrule.yaml' \
    | kopium -Af - > prometheusrule.rs
```

Example using the catalog as it were when version 1.33.6 was released:

```sh
curl -sSL 'https://raw.githubusercontent.com/CustomResourceDefinition/catalog/refs/tags/v1.33.6/definitions/monitoring.coreos.com/prometheusrule.yaml' \
    | kopium -Af - > prometheusrule.rs
```

## Adding missing CRDs

You can use the [how to contribute guide](docs/HOW-TO-CONTRIBUTE.md) to make a pull request.

Otherwise you are welcome to create an issue asking for a CRD that is not yet present in the catalog. For the best result you should include the following information in your issue, if you have it:

- Name, like `applications.argoproj.io`
- Homepage, like [https://argo-cd.readthedocs.io/en/stable/](https://argo-cd.readthedocs.io/en/stable/)
- Helm repository, like [https://argoproj.github.io/argo-helm](https://argoproj.github.io/argo-helm)
- Source repository, like [https://github.com/argoproj/argo-cd](https://github.com/argoproj/argo-cd)

If you are missing most of the details above, try to include a link to an installation guide or a short explanation of the purpose of the CRD.

## Origins

This catalog is inspired by the [CRDs-catalog](https://github.com/datreeio/CRDs-catalog) from [Datree](https://www.datree.io), but instead uses a sources list to pull updates automatically and enables re-creation from these sources provided the source helm charts, uris, etc are still available.

See a [comparison](docs/COMPARISON.md) of the schema files available in each respective repository.
