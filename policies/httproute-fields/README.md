# Description
This policy enforces specific fields for Gateway API HTTPRoute rosources based on defined values in the parameter.

# Policy logic
Currently supported:
* when `spec.allowedHostnames` list of the parameter custom resource exists, the `spec.hostnames` field of the HTTPRoute
  is defined and it only contains hostnames that are on the list

When there is no parameter custom resource the policy denys.

# Parameter used by the policy
The policy is using a mandatory custom resource (CR) kind called `VAPLibHTTPRouteFieldsParam`.

# Example parameter
```
apiVersion: vap-library.com/v1beta1
kind: VAPLibHTTPRouteFieldsParam
metadata:
  name: example
  namespace: example
spec:
  allowedHostnames:
  - test.example.com
  - test2.exmaple.com
```
