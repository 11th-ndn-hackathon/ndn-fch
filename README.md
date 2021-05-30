# NDN-FCH 2021

NDN-FCH 2021 is an online service for finding available Named Data Networking routers near the client on the global NDN testbed and other connected networks.
This project started as **NDN-FCH: The Big Rewrite™** at [11th NDN hackathon](https://11th-ndn-hackathon.named-data.net), and continues to be developed and maintained.

## API

To find available NDN routers, send an HTTP GET request to `https://fch.ndn.today/`.

Query parameters:

* **cap**: transport protocol.
  * Acceptable values: `udp`, `wss`, or `http3`.
  * Default is `udp`.
  * This is repeatable.
* **k**: number of routers.
  * If this appears once, the setting applies to every transport protocol.
  * If this is repeated, each setting applies to successive transport protocols.
* **ipv4**: `1` to accept IPv4 routers, `0` to reject IPv6 routers.
  * Default is `1`. Set to `0` if client does not have IPv4 connectivity.
* **ipv6**: `1` to accept IPv6 routers, `0` to reject IPv6 routers.
  * Default is `1` if the request is received over IPv6, otherwise `0`.
* **lon** and **lat**: client position.
  * Default is IP geolocation.

Response format:

* Text response (default) is a comma-separated list of router hostnames.
  * This format is compatible with [NDN-FCH 2016](https://github.com/named-data/ndn-fch) in most cases.
  * It is not recommended to specify multiple transport protocols in the query.
* JSON response contains host:port (for UDP) or URI (for WebSocket and HTTP/3).
  * To receive JSON response, set `Accept: application/json` request header.

## Software Components

NDN-FCH 2021 contains the following components:

* [frontend](https://github.com/11th-ndn-hackathon/ndn-fch-worker)
* [API service](https://github.com/11th-ndn-hackathon/ndn-fch)
* [health probe for UDP and WebSockets](https://github.com/eric135/ndn-reachability)
* [health probe for HTTP/3](https://github.com/yoursunny/NDN-QUIC-gateway)
