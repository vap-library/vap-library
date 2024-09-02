# Description
This policy enforces that specific resource limit types are defined for containers according to the types specified in the parameter.

# Policy logic
When `spec.enforcedResourceLimitTypes` exists in the parameter: all elements in this list must also be defined in the `resources.limits` section of all containers (`spec.containers[\*]` and
`spec.initContainers[\*]`).

Supported limit types are currently:
* cpu
* memory
* ephemeral-storage

The value provided for each specified limit is not considered; the limit type must simply be present.

Note that ephemeral containers are not considered in this policy, as since ephemeral containers are run in existing Pods and Pod resource allocations are immutable, setting `resources` on ephemeral containers is disallowed.

When there is no parameter custom resource the policy denys.

# Parameter used by the policy
The policy is using a mandatory custom resource (CR) kind called `VAPLibResourceLimitTypesParam`.

# Example parameter
```
apiVersion: vap-library.com/v1beta1
kind: VAPLibResourceLimitTypesParam
metadata:
  name: example
  namespace: example
spec:
  enforcedResourceLimitTypes:
  - cpu
  - memory
  - ephemeral-storage
```