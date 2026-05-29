---
name: source
description: Finding missing primary sources for CRDs (custom resource definitions). Use when needing to find a new primary source.
---

## Context

Kubernetes API can be extended using CRDs (custom resource definitions) - these are imported into kubernetes clusters to enable the use of CRs (custom resources). The CRs and their derived openapi specifications are collected from primary sources and stored in this repository continuously.

## Primary Source

A primary source for CRDs is published by the original authors and is available as a versioned online resource.

A primary source is not:

- extracting from a cluster.
- a similar catalog-style repository like this current one.

The primary sources are added to the configuration file, see details in [how to contribute file](../../../docs/HOW-TO-CONTRIBUTE.md).

## Instructions

You are exclusively looking for information online when locating a primary source.

You must never look at local files for any reason other than to tell if a primary source is already registered, or if certain CRDs exists. You will never read the code base.

You should suggest adding the CRD being searched for to the ignore list, if you find evidence that this CRD is:

- used exclusively as a part of testing by the original authors.
- part of an accidentally incorrectly named release.
- otherwise not meant for public consumption.

Find a primary source for this missing CRD and present your findings to the user.
