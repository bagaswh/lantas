
![Lantas?](docs/assets/lantas.png "Lantas?")

# Lantas

TCP reverse proxy with support for stream transformation through middlewares.

## Why?

It solves a specific use case of mine. I want to create a some sort of "tunnel" between two cloud network, and the data that's sent between tunnels are applied some transformations.

Application from network A wants to send data to network B. The data goes through the proxy (tunnel) in network A. 

Proxy A applies any transformations (such as compressing request) and sends the transformed data to proxy in network B.
Proxy in network B does transformations (such as decompressing request) and forwards the transformed data to the application in network B.

```   
                                    
                        +-------------+ tls & compressed  +--------------+
+-------------+         |             | ----------------> |              |        +--------------+
| client-left | ------> | lantas-left |                   | lantas-right | <----- | client-right |
+-------------+         |             | <---------------- |              |        +--------------+
                        +-------------+                   +--------------+
```

For compresion functionality to work, Lantas needs to be deployed on both sides since it sends
Lantas-specific binary format that packs compression algorithm used.

## Example configuration

- [lantas-left.yml](./examples/lantas-left.yml)
- [lantas-right.yml](./examples/lantas-right.yml)

## TODO

- [ ] Implement the binary protocol to make it actually work

## Roadmap

- [ ] Connection draining for clients and upstreams
- [ ] Upstream queueing