# Codebase Map

## Overview

CRD Catalog gathers Kubernetes CustomResourceDefinition schemas from configured sources and produces schema and definition files for downstream validation and tooling.

## Version ordering

- Values-file versions are sorted as semantic versions; prefixes, suffixes, and leading zeroes can produce incorrect ordering.

## Git sources

- Git sources are cloned with the Git CLI because cloning with the Go Git library used excessive memory.

## Local updater check

- `TestCheckLocal` is intended for step-debugging a local check and should be skipped during normal unit tests.
