# Description
This Validating Admission Policy ensures that containers drop the 'ALL' capability (i.e. removing all capabilities), and
can then only add back the 'NET_BIND_SERVICE' capability (if desired).
This policy is part of the Pod Security Standards provided by Kubernetes, found here - https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted.

# Policy logic
This policy evaluates `securityContext.capabilities.drop[]` and `securityContext.capabilities.add[]` within
`spec.containers[\*]`, `spec.initContainers[\*]` and `spec.ephemeralContainers[\*]`. For the API call to be accepted
then for each and every container: `securityContext.capabilities.drop[]` must be specified and include 'ALL', and
`securityContext.capabilities.add[]` can only include 'NET_BIND_SERVICE' (or be unspecified).

# Parameter used by the policy
This policy does not use parameters. Rules that are outlined by the PSS Restricted profile are enforced. 

# Examples
### Pass
Pass as the ALL capability is dropped.
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
      capabilities:
        drop:
        - ALL
```
### Pass
Pass as the ALL capability is dropped and NET_BIND_SERVICE is the only capability being re-added.
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
      capabilities:
        drop:
        - ALL
        add:
        - NET_BIND_SERVICE
```
### Fail
Failure as the ALL capability is not dropped.
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
Failure as a disallowed capability is added.
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
      capabilities:
        drop:
        - ALL
        add:
        - NOT_ALLOWED
```