This is a Reverse Proxy working on Layer 4.

# Protocols

1. TCP
2. UDP
3. Unix

## Supported Detail

A: TCP, Unix

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
