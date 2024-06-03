# Description
This Validating Admission Policy ensures that any defined volumes can only be of one of the allowed types.
This policy is part of the Pod Security Standards provided by Kubernetes, found here - https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted.

# Policy logic
This policy evaluates `spec.volumes[\*]`. For the API call to be accepted then for each and every volume: one of the following fields must be set to a non-null value:

* `spec.volumes[*].configMap`
* `spec.volumes[*].csi`
* `spec.volumes[*].downwardAPI`
* `spec.volumes[*].emptyDir`
* `spec.volumes[*].ephemeral`
* `spec.volumes[*].persistentVolumeClaim`
* `spec.volumes[*].projected`
* `spec.volumes[*].secret`

# Parameter used by the policy
This policy does not use parameters. Rules that are outlined by the PSS Restricted profile are enforced. 

# Examples
### Pass
Pass as all volumes are of an allowed type.
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
  volumes:
  - name: example-volume
    configMap:
      name: log-config
      items:
        - key: log_level
          path: log_level
```
### Fail
Failure as a volume is of a disallowed type.
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
  volumes:
  - name: example-volume
    hostPath:
      path: /data/foo # directory location on host
      type: Directory # this field is optional
```