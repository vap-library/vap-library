# Description
This policy enforces that the HTTPRoute defines the `spec.hostnames` field and it only contains hostnames that are in
the `spec.allowedHostnames` list of the parameter custom resource. When there is no parameter custom resource the policy
denys.

# Parameter used by the policy
The policy is using a custom resource (CR) kind called `VAPLIBHTTPRouteHostnameParam`. The CR has to lists the allowed
hostnames in a list of strings field called `spec.allowedHostnames`.

# Example parameter
```
apiVersion: vap-library.io/v1beta1
kind: VAPLIBHTTPRouteHostnameParam
metadata:
  name: example
  namespace: example
spec:
  allowedHostnames:
  - test.example.com
  - test2.exmaple.com
```