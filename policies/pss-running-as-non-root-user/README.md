# Description
This Validating Admission Policy ensures that containers do not set to run as the root user **in the k8s manifest**.
(UID 0) This policy is part of the Pod Security Standards provided by Kubernetes, found here - https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted

> **_IMPORTANT:_ This policy follows what PSS defines for the `Running as Non-root user` control but this policy alone
> would not prevent a user from running a container as root! In case the container image metadata defines root and the
> `runAsUser` is not defined in the pod spec or the container spec, the container will run as root. In order to prevent
> that, make sure the pss-running-as-non-root policy is enforced!**

# Policy logic
This policy uses the `runAsUser` filed which takes in the user ID to run the container/s. For the API call to be
accepted then the parameter is required to be set to a non-zero integer or not defined. That's to say, the only case
where the request will be rejected is where `securityContext.runAsRoot: 0` is specified, either at the spec level, or
within spec.containers, spec.initContainers or spec.ephemeralContainers.

> **_IMPORTANT:_ This policy follows what PSS defines for the `Running as Non-root user` control but this policy alone
> would not prevent a user from running a container as root! In case the container image metadata defines root and the
> `runAsUser` is not defined in the pod spec or the container spec, the container will run as root. In order to prevent
> that, make sure the pss-running-as-non-root policy is enforced!**
 

# Parameter used by the policy
This policy does not use parameters.

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