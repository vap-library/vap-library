# Description
This Validating Admission Policy ensures that containers explicitly disallow privilege escalation.
This policy is part of the Pod Security Standards provided by Kubernetes, found here - https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted.

# Policy logic
This policy evaluates `securityContext.allowPrivilegeEscalation` within `spec.containers[\*]`, `spec.initContainers[\*]` and `spec.ephemeralContainers[\*]`. For the API call to be accepted then for each and every container: `securityContext.allowPrivilegeEscalation` must be set to false.

# Parameter used by the policy
This policy does not use parameters. Rules that are outlined by the PSS Restricted profile are enforced. 

# Examples
### Pass
Pass as all containers have allowPrivilegeEscalation set to false.
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
      allowPrivilegeEscalation: false
  initContainers:
  - name: init-example
    image: example
    securityContext:
      allowPrivilegeEscalation: false
```
### Fail
Failure as allowPrivilegeEscalation is not set to false on all containers.
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
      allowPrivilegeEscalation: false
  initContainers:
  - name: init-example
    image: example
```
### Fail
Failure as allowPrivilegeEscalation is not set to false on all containers.
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
      allowPrivilegeEscalation: true
```