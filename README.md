# Kubernetes Validating Admission Policy library
This repo contains community maintained (NO WARRANTY) collection of [Kubernetes Validating Admission Policies](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/). That can be used with few commands and namespace labels.

# Status
**The test framework is still receiving improvements**

# Testing

Prerequisites:
- Go v1.22.x
- Docker (for Kind)

To run all the tests: 
```bash
go test ./policies/...
```

To run tests for a single policy 
```bash
go test  ./policies/POLICYNAME/
```

# Installing and using the library
*NOTE: Validating Admission Policy is beta in 1.28+ and disabled by default in most Kubernetes distributions. Follow the [official instructions](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#before-you-begin) to enable it on your k8s cluster/distribution*

**Every CRD that is used for policy parameter has a name prefix of `VAPLib` and every resource that we create has a suffix of `.vap-library.com` to avoid name collisions. As such the resources can be safely applied from the release manifest files**

Parameter CRDs, policies and policy bindings are available in separate yaml file as [release artifacts](https://github.com/vap-library/vap-library/releases/latest)

## To apply ALL
It is possible to apply all policies, policy bindings and parameter CRDs available in the vap-library (this would not enforce anything without proper labels on the namespaces):
```
export VAPRELEASE=v0.1.1
kubectl apply -k https://github.com/vap-library/vap-library.git/release?ref=${VAPRELEASE}
```

## Enforcing a policy
Make sure that you create a parameter ConfigMap or CR in case the policy requires it. You can enforce the policy with applying the relevant label to the namespace with a `deny` value (to warn them use the `warn` value):
```
vap-library.com/POLICYNAME: deny
```

# Sources that can help for contribution
* A great repo to review and learn CEL: [ARMOS's](https://www.armosec.io/) [cel-admission-library](https://github.com/kubescape/cel-admission-library/tree/main)
* [Official VAP documentation](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
* [Kubernetes CEL documenation](https://kubernetes.io/docs/reference/using-api/cel/)
