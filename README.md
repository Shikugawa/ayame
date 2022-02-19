# ayame

A simple network laboratory builder with Linux namespaces

### Prerequisites

- iproute2
- OpenvSwitch

### Examples

Create config and save as `sample.yaml`

```
# L2 connectivity is supported only by veth and OpenvSwitch.
# All the link names must not be duplicated.
links:
  - name: veth1
    mode: direct_link # use veth
  - name: br1
    mode: bridge # use OpenvSwitch

# All the namespace names must not be duplicated.
namespaces:
  - name: ns1
    devices:
      - name: veth1 # device name must be defined in links
        cidr: 192.168.100.10/24
  - name: ns2
    devices:
      - name: veth1 # device name must be defined in links
        cidr: 192.168.100.11/24
  - name: ns3
    devices:
      - name: br1 # device name must be defined in links
        cidr: 182.102.101.11/24
  - name: ns4
    devices:
      - name: br1 # device name must be defined in links
        cidr: 182.102.101.12/24
  - name: ns5
    devices:
      - name: br1 # device name must be defined in links
        cidr: 182.102.101.13/24
```

Run `sudo ayame create -c sample.yaml`
