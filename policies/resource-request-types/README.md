# Description
This policy enforces that specific resource request types are defined for containers according to the types specified in the parameter.

# Policy logic
When `spec.enforcedResourceRequestTypes` exists in the parameter: all elements in this list must also be defined in the `resources.requests` section of all containers (`spec.containers[\*]` and
`spec.initContainers[\*]`).

Supported request types are currently:
* cpu
* memory
* ephemeral-storage

The value provided for each specified request is not considered; the request type must simply be present.

Note that ephemeral containers are not considered in this policy, as since ephemeral containers are run in existing Pods and Pod resource allocations are immutable, setting `resources` on ephemeral containers is disallowed.

When there is no parameter custom resource the policy denys.

# Parameter used by the policy
The policy is using a mandatory custom resource (CR) kind called `VAPLibResourceRequestTypesParam`.

# Example parameter
```
apiVersion: vap-library.com/v1beta1
kind: VAPLibResourceRequestTypesParam
metadata:
  name: example
  namespace: example
spec:
  enforcedResourceRequestTypes:
  - cpu
  - memory
  - ephemeral-storage
```