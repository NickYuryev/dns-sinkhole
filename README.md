# dns-sinkhole

A small DNS resolver that drops ad and malware domains before they ever leave the network.

## How it works

The server listens for UDP DNS queries, checks the queried name against a hosts-format blocklist (StevenBlack by default), and returns `NXDOMAIN` for any match. Everything else is forwarded to an upstream resolver and the response is cached using the TTL from the answer. A Bloom filter sits in front of the sorted domain list so misses cost almost nothing.

## Quick start

```
make download-blocklist
make run
```

In another shell:

```
dig @127.0.0.1 -p 8053 example.com          # resolves normally
dig @127.0.0.1 -p 8053 googleadservices.com  # NXDOMAIN
```

Metrics are at `http://127.0.0.1:9090/metrics`.

## Configuration

| Flag | Default | Purpose |
|---|---|---|
| `--port` | `8053` | UDP port for DNS |
| `--metrics-port` | `9090` | TCP port for Prometheus |
| `--upstream` | `1.1.1.1:53` | Upstream DNS resolver |
| `--blocklist` | `./blocklist.txt` | Path or `http(s)://` URL to a hosts-format blocklist |
| `--verbose` | `false` | Log every query |

## Docker

```
docker build -t dns-sinkhole .
docker run --rm -p 8053:8053/udp -p 9090:9090 dns-sinkhole
```

The image is multi-stage and ships a static binary on top of Alpine.

## Deployment

See `deploy/` for a systemd service, a nightly timer that refreshes the blocklist, and a logrotate config. Binding to port 53 as an unprivileged user needs `CAP_NET_BIND_SERVICE`, which the unit already sets.

## Testing

```
make test
make bench
```

## License

MIT
