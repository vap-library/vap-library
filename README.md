# Kubernetes Validating Admission Policy library
This repo contains UNOFFICIAL, community maintained collection of [Kubernetes Validating Admission Policies](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/).

# Using the library
*NOTE: Validating Admission Policy is beta in 1.28+ and disabled by default. Follow the [official instructions](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#before-you-begin) to enable it on your k8s cluster/distribution*

**Every policy (and related parameter CRD if exists) has a name prefix of `vap-library-`. This highly reduces the risk of name collisions when the library gets installed.**
## Install using kubectl
### parameter CRDs
Some policies that require complex parameters (that cannot be easily represented in a parameter ConfigMap) has their own parameter CRDs. Clone the repo with the selected `ref` and use following command to install the parameter CRDs
```
kubectl install install/vap-library-parameter-crds.yaml
```

### policies
```
kubectl install install/vap-library-policies.yaml
```

## Create resources for parameters (if needed)
Most of the policies require parameters which could be either a `ConfigMap` or a `Custom Resource`. Check the tests in the policy's directory for exmaples.

## Create policy binding
Create the [ValidatingAdmissionPolicyBinding](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#what-resources-make-a-policy) to bind the policy (and reference the parameter) to selected resources.

# Sources
* Many of the policies were based on [Kubescape's](https://www.armosec.io/kubescape/) [cel-admission-library](https://github.com/kubescape/cel-admission-library/tree/main). Great repo to review and learn from
* [Official VAP documentation](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
* [Kubernetes CEL documenation](https://kubernetes.io/docs/reference/using-api/cel/)

# NO WARRANTY INCLUDED
**While the policies have been written and/or reviewed by experienced k8s engineers/developers this is a community driven, public repository. Use it at your own risk!**
