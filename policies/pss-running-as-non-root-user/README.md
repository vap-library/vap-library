# Description
This Validating Admission Policy ensures that containers are not run as the root user. (UID 0)
This policy is part of the Pod Security Standards provided by Kubernetes, found here - https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted

# Parameter used by the policy
This policy uses the `runAsUser` parameter which takes in the user ID to run the container/s. For the API call to be accepted then the parameter is required to be set to a non-zero integer or not defined.
That's to say, the only case where the request will be rejected is where `securityContext.runAsRoot: 0` is specified, either at the spec level, or within spec.containers, spec.initContainers or spec.ephemeralContainers.

# Example parameter
### Pass
Pass due to no runAsUser parameter being defined.
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
### Pass
Pass as the runAsUser parameter isn't set to 0.
```
apiVersion: v1
kind: Pod
metadata:
  name: example
  namespace: example
spec:
  securityContext:
    runAsUser: 1000
  containers:
  - name: example
    image: example
```
### Pass
Pass as the runAsUser parameter isn't set to 0.
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
      runAsUser: 1000
```
### Pass
Pass as all runAsUser parameters aren't set to 0.
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
      runAsUser: 1000
  - name: example
    image: example
    securityContext:
      runAsUser: 1000
```
### Fail
Failure due to runAsUser being set to 0.
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
      runAsUser: 0
```