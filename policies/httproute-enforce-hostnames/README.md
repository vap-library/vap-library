# Description
This policy enforces that the HTTPRoute defines the `spec.hostnames` field and it only contains hostnames that are in
the `spec.allowedHostnames` list of the parameter custom resource. When there is no parameter custom resource the policy
denys.

# Parameter used by the policy
The policy is using a mandatory custom resource (CR) kind called `HTTPRouteEnforceHostnamesParam`. The CR has to list the allowed
hostnames in an array of strings field called `spec.allowedHostnames`.

# Example parameter
```
apiVersion: vap-library.com/v1beta1
kind: VAPLIBHTTPRouteHostnameParam
metadata:
  name: example
  namespace: example
spec:
  allowedHostnames:
  - test.example.com
  - test2.exmaple.com
```
