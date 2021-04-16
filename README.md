# ayame

A simple network laboratory builder with Linux namespaces

### Prerequisites

- iproute2

### Examples

Create config and save as `sample.yaml`

```
veth:
  - left: veth1
    right: veth1-target

namespace:
  - name: ns1
    device:
      - name: veth1
        cidr: 192.168.100.10/24
```

Run `ayame create -c sample.yaml`
