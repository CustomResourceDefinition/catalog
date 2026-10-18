---
name: source-audit
description: Audit configured sources that might be stale. Use when needing to verify a source is still active.
---

## Context

Kubernetes API can be extended using CRDs (custom resource definitions) - these are imported into kubernetes clusters to enable the use of CRs (custom resources). The CRs and their derived openapi specifications are collected from primary sources and stored in this repository continuously.

# Configured Source

Sources are added to the configuration file, see details in [how to contribute file](../../../docs/HOW-TO-CONTRIBUTE.md).

Sources can have a description field which may describe that a source is already deprecated, archived or otherwise unavailable. There is no need to re-verified a source with such a description.

Sources with `skipped: true` are intentionally excluded from being used for generating new content. Confirm they should still be skipped based on explanation in the description.

You can look up a configuration entry with its (sops used in the example) with:

```sh
yq '.[] | select(.name == "sops")' configuration.yaml
```

# Registry

There is a registry file at `registry.yaml` which contains among other details the last updated field for each source.

The output list corresponds to configuration names. The registry key is `{name}` when `entries` is omitted or empty, and `{name}.{entry}` for each entry whenever `entries` is nonempty (including a single entry).

You can list the oldest 100 updates from active sources with:

```sh
if ! yq --version 2>/dev/null |
  grep -Eq 'github\.com/mikefarah/yq/.*version v4\.'; then
  printf '%s\n' 'Requires Mike Farah yq v4.' >&2
  exit 1
fi

yq '
  [load("configuration.yaml")[] | select(.skipped != true)
    | .name as $name | (.entries // [""])[]
    | [$name, .] | map(select(. != "")) | join(".")
  ] as $active
  | .sources
  | to_entries
  | map(select(.key as $key | $active | any_c(. == $key)))
  | sort_by(.value.lastUpdated)
  | .[:100]
  | group_by(.value.kind)
  | .[]
  | {"kind": .[0].value.kind, "sources": map(.key)}
' registry.yaml
```

You must use the above command to find sources to audit, if no target sources are given by the user.

# Instructions

For Git Sources:

- Visit the repo
- Look for archive banners, redirects, 404s, or explicit deprecation notices
- Check if the repo is active (recent releases, commits)

For Helm and OCI sources:

- Add repo to helm
- List charts at helm repo
- Verify configured source charts are available
- Remove repo

For http sources and when it is unclear if a git/oci/helm source is available, you should copy the singular configuration entry to `build/configuration.yaml` and run `make check`.

Suggest unavailable sources have their description updated to reflect their state, if it is not already.

Suggest adding or finding a primary sources if you find evidence that the source has been replaced or superseded.

There is no need to ask questions - do the work, you are finished only once you have:

- found the state of all audit targets
- organize your findings by kind
- presented a finished report to the user
