# ecs_remap

Coredns plugin to change the EDNS client subnet based on an IP map.

The primary usecase is to remap private IP addresses from a corporate VPN to their real public IP address, to be passed to the upstream DNS server for better CDN geolocation.  

Usage:

Add the following to `plugins.cfg`, before `debug`:

```
ecs_remap:github.com/nicelocal/ecs_remap
```

And use the following example Corefile:
```
. {
    ecs_remap {
        192.168.1.2 1.2.3.0/24
        192.168.1.3 4.3.2.0/24

        ff::1 4.3.2.0/24
        192.168.69.4 2001:db8::/32
    }

    # Do not use 1.1.1.1, it explicitly blocks ECS
    forward . 8.8.8.8
}
```

Example dockerfile:

```
FROM golang

RUN git clone -b v1.12.0 --depth 1 https://github.com/coredns/coredns /coredns && \
    cd /coredns && \
    sed '/bind:bind/a ecs_remap:github.com/nicelocal/ecs_remap' plugin.cfg -i && \
    make

FROM scratch

COPY --from=0 /coredns/coredns /coredns

ENTRYPOINT ["/coredns"]
```