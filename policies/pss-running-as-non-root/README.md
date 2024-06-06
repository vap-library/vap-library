# Description
This Validating Admission Policy ensures that containers are run as non-root users.
This policy is part of the Pod Security Standards provided by Kubernetes, found here - https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted.

# Policy logic
This policy evaluates securityContext.runAsNonRoot within spec (Pod-level), `spec.containers[\*]`,
`spec.initContainers[\*]` and `spec.ephemeralContainers[\*]`. For the API call to be accepted then for each and every
container: `securityContext.runAsNonRoot` must be set to true. Note that the value defined at the container level takes
precedence over that defined at Pod-level.

# Parameter used by the policy
This policy does not use parameters.

# Examples
### Pass
Pass as all containers have runAsNonRoot set to true.
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
      runAsNonRoot: true
  initContainers:
  - name: init-example
    image: example
    securityContext:
      runAsNonRoot: true
```
### Pass
Pass as runAsNonRoot is set to true at the Pod level, and this is not contradicted in any containers.
```
apiVersion: v1
kind: Pod
metadata:
  name: example
  namespace: example
spec:
  securityContext:
    runAsNonRoot: true
  containers:
  - name: example
    image: example
  initContainers:
  - name: init-example
    image: example
```
### Fail
Failure as runAsNonRoot is not set to true on all containers, and it is not set to true at the Pod level.
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
      runAsNonRoot: true
  initContainers:
  - name: init-example
    image: example
```
### Fail
Failure as although runAsNonRoot is set to true at the Pod level, the is overridden by a container-level parameter.
```
apiVersion: v1
kind: Pod
metadata:
  name: example
  namespace: example
spec:
  securityContext:
    runAsNonRoot: true
  containers:
  - name: example
    image: example
    securityContext:
      runAsNonRoot: false
```