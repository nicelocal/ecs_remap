# ecs_remap

Coredns plugin to change the EDNS client subnet based on an IP map.

The primary usecase is to remap private IP addresses from a corporate VPN to their real public IP address, to be passed to the upstream DNS server for better CDN geolocation.  

Usage:

```
. {
    ecs_remap {
        192.168.1.2 1.2.3.0/24
        192.168.1.3 4.3.2.0/24
    }
}
```