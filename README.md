# Kubernetes Validating Admission Policy library
This repo contains a community maintained collection of [Kubernetes Validating Admission Policies](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
and a **testing framework** that can be used to verify that admission policies are doing what they intended to do.

**The policies in the library can be installed with a few commands and can be enforced with namespace labels.**

# Installing and using the library
Validating Admission Policy (VAP) has been promoted to GA in Kubernetes 1.30. The policies in this library are using the
v1 API and as such, they **require Kubernetes 1.30 or newer**.

> **_NOTE:_** Validating Admission Policy was beta in 1.28 and 1.29 and were disabled by default in most Kubernetes
> distributions. One could modify the API of the policies from `v1` to `v1beta1` and most probably they would work in
> those version of Kubernetes as well. For 1.28 and 1.29 follow the [official instructions](https://v1-29.docs.kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#before-you-begin)
> to enable VAP on your k8s cluster/distribution*. However, we do not run tests on older K8s versions.

Every CRD that is used for policy parameter has a name prefix of `VAPLib` and every resource that the library creates
has a suffix of `.vap-library.com` to avoid name collisions. This allows that the resources can be safely applied from
the release manifest files on existing clusters.

Parameter CRDs, policies and policy bindings are available in separate yaml files as [release artifacts](https://github.com/vap-library/vap-library/releases/latest)

## To apply ALL
It is possible to apply all policies, policy bindings and parameter CRDs available in the vap-library (this would not
enforce anything without proper labels on the namespaces):
```
export VAPRELEASE=v0.1.9
kubectl apply -k https://github.com/vap-library/vap-library.git/release-process/release?ref=${VAPRELEASE}
```

## To apply SELECTED
To not enforce a certain policy, one can simply not add the given label (specified in the binding) to the namespace.

In addition, it is possible to generate a custom subset of the policies, policy bindings, and parameter CRDs available in the vap-library. To do this, a release script exists which takes a yaml config file, and generates custom release artifacts (`policies.yaml`, `bindings.yaml`, `crds.yaml`) based on the provided config.

Everything associated with the release process sits in the `release-process` directory. In order to create a custom release:
1) Create a new config file specifying the desired policies and bindings. The file `release-process/full-release-config.yaml` will include all policies from the library, along with a pair of bindings (deny+audit & warn) for each, and all CRDs, so this should be used as a template, removing/modifying any entries as desired
2) Run the script, providing the path to the prepared config file, e.g: `python release.py full-release-config.yaml`

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
go clean -testcache && go test -p 2 ./policies/...
```

To run tests for a single policy (use -v for verbose output):
```bash
go clean -testcache && go test  ./policies/POLICYNAME/
```

> **_NOTE:_** in case test fails with error it may leak kind cluster. Cleanup with `kind delete clusters --all`

## Maintainers
Versioned release artifacts are generated automatically by the GitHub action defined in `.github/workflows/release.yaml`. The full config and generated release artifacts found in `release-process` should always represent the complete set of policies available in the repository, with `Deny&Audit` and `Warn` bindings for each policy. 

To generate a new release, the process is as follows:
1) Create a new feature branch
2) Ensure any new policy is defined in the standard way, with a folder under `policies` containing any required CRD parameters, the policy itself, a set of tests, and a README
3) Update `release-process/full-release-config.yaml` with a new section for any new policy, and the two bindings (use other config entries as examples, they will be very similar)
4) If desired, run the release script (`release.py`) locally as per the instructions above
5) Bump the version found in `release-process/version`, as per semantic versioning
6) Update the Policies table above in this README, adding details of any new policies
7) Push your changes, and submit a pull request
8) Once approved and merged, the GitHub Action will run, automatically running the script, thus overwriting the output files found in `release-process/release` and creating a new release artifact, named as per the semantic version in `release-process/version`.

# Sources that can help for contribution
* [Official VAP documentation](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
* [Kubernetes CEL documenation](https://kubernetes.io/docs/reference/using-api/cel/)
* A repo that inspired us: [ARMOS's](https://www.armosec.io/) [cel-admission-library](https://github.com/kubescape/cel-admission-library/tree/main)

