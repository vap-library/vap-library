# Kubernetes Validating Admission Policy library
This repo contains a community maintained collection of [Kubernetes Validating Admission Policies](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
and a **testing framework** that can be used to verify that admission policies are doing what they intended to do.

**The policies in the library can be installed with a few commands and can be enforced with namespace labels.**

# Installing and using the library
> **_NOTE:_** Validating Admission Policy is beta in 1.28+ and disabled by default in most Kubernetes distributions up
> to 1.30 (in which it turned to GA and got enabled by default). Follow the [official instructions](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#before-you-begin)
> to enable it on your k8s cluster/distribution*

Every CRD that is used for policy parameter has a name prefix of `VAPLib` and every resource that the library creates
has a suffix of `.vap-library.com` to avoid name collisions. This allows that the resources can be safely applied from
the release manifest files on existing clusters.

Parameter CRDs, policies and policy bindings are available in separate yaml files as [release artifacts](https://github.com/vap-library/vap-library/releases/latest)

## To apply ALL
It is possible to apply all policies, policy bindings and parameter CRDs available in the vap-library (this would not
enforce anything without proper labels on the namespaces):
```
export VAPRELEASE=v0.1.7
kubectl apply -k https://github.com/vap-library/vap-library.git/release-process/release?ref=${VAPRELEASE}
```

## To apply SELECTED
It is possible to generate a custom subset of the policies, policy bindings, and parameter CRDs available in the vap-library. To do this, a release script exists which takes a yaml config file, and generates custom release artifacts (`policies.yaml`, `bindings.yaml`, `crds.yaml`) based on the provided config.

Everything associated with the release process sits in the release-process directory. In order to create a release:
1) Create a new config file specifying the desired policies and bindings. The file `release-process/config/example-release-config.yaml` will include all policies from the library, along with a pair of bindings (deny+audit & warn) for each, and all CRDs, so this should be used as a template, removing/modifying any entries as desired.
2) Run the script, providing the path to the prepared config file: `python release.py ../config/example-release-config.yaml`

For the release script to correctly include a policy and associated resources, the policy must be in its own directory under `./policies`, and any CRD must be in the same directory, named `crd-parameter.yaml`. See existing policies for reference.

The generated yaml files can then be applied. As with applying ALL, note that the proper labels must be set on the namespaces in order for the policies to enforce anything.

## Enforcing a policy
Make sure that you create a parameter ConfigMap or CR in case the policy requires it. You can enforce the policy with
applying the relevant label to the namespace with a `deny` value (to warn them use the `warn` value):
```
vap-library.com/POLICYNAME: deny
```

# Policies
| Policy name                  | Description                                                                                                                                                                                                                                                               | Parameter                                    |
|------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------|
| pss-capabilities             | Enforces container capabilities as outlined by the [Pod Security Standard restricted profile](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted).                                                                                           | N/A                                          |
| pss-privilege-escalation     | Ensures that containers explicitly disallow privilege escalation as outlined by the [Pod Security Standard restricted profile](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted).                                                          | N/A                                          |
| pss-running-as-non-root      | Ensures that containers are run as non-root users as outlined by the [Pod Security Standard restricted profile](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted).                                                                         | N/A                                          |
| pss-running-as-non-root-user | Ensures that containers do not set to run as the root user **in the k8s manifest** as outlined by the [Pod Security Standard restricted profile](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted).                                        | N/A                                          |
| pss-seccomp                  | Ensures that containers explicitly set the Seccomp profile to one of the allowed values (`RuntimeDefault` or `Localhost`) as outlined by the [Pod Security Standard restricted profile](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted). | N/A                                          |
| pss-volume-types             | Ensures that any defined volumes can only be of one of the allowed types as outlined by the [Pod Security Standard restricted profile](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted).                                                  | N/A                                          |
| service-type                 | Ensures that `Service` resources can only use types that are listed in the `spec.allowedTypes` field of the parameter.                                                                                                                                                    | `VAPLibServiceTypeParam` (Mandatory)         |
| httproute-fields             | Ensures that specific fields of `HTTPRoute` resources match the defined values from parameter.                                                                                                                                                                            | `VAPLibHTTPRouteFieldsParam` (Mandatory)  |
| kustomization-fields         | Ensures that specific fields of [Flux Kustomization](https://fluxcd.io/flux/components/kustomize/kustomizations/) resources match defined values from parameter.                                                                                                          | `VAPLibKustomizationFieldsParam` (Mandatory) |
| helmrelease-fields           | Ensures that specific fields of [HelmRelease](https://fluxcd.io/flux/components/helm/helmreleases/) resources match defined values from parameter.                                                                                                                        | `VAPLibHelmReleaseFieldsParam` (Mandatory)   |

# Testing of the policies
A "testing framework" has been developed (based on Kubernetes e2e) to support testing of admission policies.

Prerequisites:
- Go v1.22.x
- Docker (for Kind)

To run all the tests (use -v for verbose output): 
```bash
go test -p 2 ./policies/...
```

To run tests for a single policy (use -v for verbose output):
```bash
go test  ./policies/POLICYNAME/
```

To test the yaml for a release can be applied to a cluster without error (use -v for verbose output):
```bash
go test ./release-process/release/
```

# Sources that can help for contribution
* [Official VAP documentation](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
* [Kubernetes CEL documenation](https://kubernetes.io/docs/reference/using-api/cel/)
* A repo that inspired us: [ARMOS's](https://www.armosec.io/) [cel-admission-library](https://github.com/kubescape/cel-admission-library/tree/main)

