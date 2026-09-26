# CRD Catalog

This domain covers the collection of Kubernetes CRDs and their schemas from external sources for use by validation and development tools.

## Language

**Source**:
A configured route for obtaining CRDs from an upstream.
_Avoid_: Provider, feed

**Upstream**:
The external repository, chart repository, or service from which a source obtains CRDs.
_Avoid_: Source

**Source version**:
A release or revision of an upstream from which CRDs are obtained.
_Avoid_: CRD version

**CRD**:
A Kubernetes CustomResourceDefinition that establishes a custom resource kind within an API group.
_Avoid_: CRD definition

**CRD version**:
An API version declared by a CRD.
_Avoid_: Source version

**Schema**:
The validation schema associated with a CRD version; a CRD version may have no schema.
_Avoid_: CRD

## Relationships

- A **Source** obtains CRDs from an **Upstream** at a **Source version**.
- A **Source** can provide multiple **CRDs**.
- A **CRD** declares one or more **CRD versions**.
- A **Schema** belongs to a **CRD version**, and a **CRD version** may have no **Schema**.
