# Description
This Validating Admission Policy ensures that containers explicitly set the Seccomp profile to one of the allowed values (RuntimeDefault or Localhost). Both the Unconfined profile and the absence of a profile are prohibited.
This policy is part of the Pod Security Standards provided by Kubernetes, found here - https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted.

# Policy logic
This policy evaluates `securityContext.seccompProfile.type` within `spec` (Pod-level), `spec.containers[\*]`, `spec.initContainers[\*]` and `spec.ephemeralContainers[\*]`. For the API call to be accepted then for each and every container: `securityContext.seccompProfile.type` must be set to `RuntimeDefault` or `Localhost`, unless it is set at the Pod level. If at the Pod-level or in any individual container the parameter is set to a value other than `RuntimeDefault` or `Localhost`, the request is rejected.

# Parameter used by the policy
This policy does not use parameters. Rules that are outlined by the PSS Restricted profile are enforced. 

# Examples
### Pass
Pass as all containers have an allowed value.
```
apiVersion: v1
kind: Pod
metadata:
  name: example
  namespace: example
spec:
  containers:
  - name: example
    image: example
    securityContext:
      seccompProfile:
        type: Localhost
```
### Pass
Pass as the parameter is set to an allowed value at Pod-level, and no containers contradict this.
```
apiVersion: v1
kind: Pod
metadata:
  name: example
  namespace: example
spec:
  securityContext:
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: example
    image: example
```
### Fail
Failure as the parameter is not set to an allowed value.
```
apiVersion: v1
kind: Pod
metadata:
  name: example
  namespace: example
spec:
  containers:
  - name: example
    image: example
```
### Fail
Failure as the parameter is set to a disallowed value on one or more of the containers.
```
apiVersion: v1
kind: Pod
metadata:
  name: example
  namespace: example
spec:
  securityContext:
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: example
    image: example
    securityContext:
      seccompProfile:
        type: Unconfined
```