This is a Reverse Proxy working on Layer 4. On Production, recommend [HAProxy](https://github.com/haproxy/haproxy), which work very well on Layer 4, also support Unix & TLS.

# Protocols

1. TCP
2. TLS
3. UDP
4. Unix

## Supported Detail

A: TLS, TCP, Unix

B: UDP

| Inbound | Outbound | Support |
| -- | -- | -- |
| A | A | Y |
| A | B | N |
| B | A | Y |
| B | B | Y |



# Usage
```
./proxysocket inbound_arguemnt outbound_arguemnt
./proxysocket udp://0.0.0.0:30053 unix:///var/run/dns.socket
./proxysocket unix:///var/run/dns.socket udp://127.0.0.1:53
```
