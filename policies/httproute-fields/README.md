# Description
This policy enforces specific fields for Gateway API HTTPRoute rosources based on defined values in the parameter.

# Policy logic
Currently supported:
* when `spec.allowedHostnames` exists in the parameter: the `spec.hostnames` field of the HTTPRoute must be defined
  and can only contain hostnames that are on the parameter list
* when `spec.allowedParentRefs` exists in the parameter: if any of the allowedParentRef list item matches with ALL it's
  attributes on the requested HTTPRoute resource, then the resource is allowed otherwise it is rejected 


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
  allowedParentRefs:
  - name: name-only-gateway
  - name: with-namespace-gateway
    namespace: gateway-namespace

```
