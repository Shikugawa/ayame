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
    commands: # run commands inside namespaces
      - sysctl -w net.ipv4.ip_forward=1
      # it supports variables in the command definition.
      # Variables should be used as the following format: `$(DEVICE_NAME)`
      # DEVICE_NAME must be defined in the devices. In this example, we can use only `veth1` as a variable.
      - iptables -A FORWARD -i $(veth1) -d 10.0.0.1 -j ACCEPT
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
