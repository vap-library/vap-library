# Kubernetes Validating Admission Policy library
This repo contains community maintained collection of [Kubernetes Validating Admission Policies](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/).

# Status
**The directory structure and the test framework is under heavy development**

# Testing

Prerequisites:
- Go v1.22.x
- kubectl v1.29.x
- kind v0.22.x

To run all the tests: 
```bash
go test ./...
```

To run tests for a single policy cd into that policy's directory and run 
```bash
go test
```

# Releasing

In order to release a new version of the library, just create a tag and push it.
```
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

# Using the library
*NOTE: Validating Admission Policy is beta in 1.28+ and disabled by default in most Kubernetes distributions. Follow the [official instructions](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#before-you-begin) to enable it on your k8s cluster/distribution*

**Every policy (and related parameter CRD if exists) has a name prefix of `vap-library.com` to avoid name collisions.**

## Download the policies
Go to https://github.com/vap-library/vap-library/releases/latest and download the `policies.yaml`, `bindings.yaml` and `crds.yaml` files.


## Install using kubectl
These files contain all the policies, policy bindings and CRDs for parameters respectively. If you want to exclude a particular policy, you can remove it from the `bindings.yaml` file.

### parameter CRDs
Some policies that require complex parameters (that cannot be easily represented in a ConfigMap) has their own CRDs for parameters. Clone the repo with the selected `ref` and use following command to install the CRDs
```
kubectl apply -f install/vap-library-parameter-crds.yaml
```

### policies
```
kubectl apply -f install/vap-library-policies.yaml
```

## Create resources for parameters (if needed)
Most of the policies require parameters which could be either a `ConfigMap` or a `Custom Resource`. Check the tests in the policy's directory for examples.

## Create policy binding
Create the [ValidatingAdmissionPolicyBinding](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#what-resources-make-a-policy) to bind the policy (and reference the parameter) to selected resources.

# Sources
* We are planning to port some of the policies that are available in [ARMOS's](https://www.armosec.io/) [cel-admission-library](https://github.com/kubescape/cel-admission-library/tree/main). Great repo to review and learn from
* [Official VAP documentation](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
* [Kubernetes CEL documenation](https://kubernetes.io/docs/reference/using-api/cel/)

# NO WARRANTY INCLUDED
**REMINDER: These policies are maintained by the k8s community. Use it at your own risk!**
