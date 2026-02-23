# Validating Admission Policies Library
## v0.1.12
* add no-default-sa-rolebinding policy
* update vendored dependencies (gateway-api v1.4.1, flux kustomize-controller v1.8.0, flux helm-controller v1.5.0)
* update Go dependencies to latest stable
* update HelmRelease test apiVersion to helm.toolkit.fluxcd.io/v2

## v0.1.11
* no change in policies but tests were also run against K8s versions 1.31 and 1.32

## v0.1.10
* add resource-limit-types and resource-request-types policies

## v0.1.9
* remove obsolete bindings from the policies
