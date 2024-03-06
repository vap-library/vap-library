# Description
This policy enforces that Service resources can only use allowed types that are in the `spec.allowedTypes` list of the
parameter custom resource. When there is no parameter custom resource the policy denys.

# Parameter used by the policy
The policy is using a mandatory custom resource (CR) kind called `VAPLibServiceEnforceTypeParam`. The CR has to list the
allowed types in an array of strings field called `spec.allowedTypes`.

# Example parameter
```
apiVersion: vap-library.com/v1beta1
kind: VAPLibServiceEnforceTypeParam
metadata:
  name: service-enforce-type.vap-library.com
  namespace: test
spec:
  allowedTypes:
  - ClusterIP
  - NodePort
```
