# Benchmark harness

Everything behind the post *Cloudflare's CORS proxy Worker, benchmarked against
corsproxy* (corsproxy.dev/blog/cloudflare-cors-proxy-worker-benchmark). One
Docker network, one upstream, two proxies, Apache Bench, seven probes.

`worker/src/index.js` is the JavaScript tab of Cloudflare's
[CORS header proxy](https://developers.cloudflare.com/workers/examples/cors-header-proxy/)
example, unmodified. `worker/config.capnp` runs it on bare `workerd`.

```bash
docker network create bench
docker build -t bench-upstream upstream
docker build -t cf-workerd worker

docker run -d --name up  --network bench bench-upstream
docker run -d --name px  --network bench -e ALLOW_PRIVATE_NETWORKS=true ghcr.io/melihbirim/corsproxy:1.2.6
docker run -d --name px2 --network bench ghcr.io/melihbirim/corsproxy:1.2.6
docker run -d --name px3 --network bench -e ALLOW_PRIVATE_NETWORKS=true -e ALLOWED_ORIGINS=https://app.example.com ghcr.io/melihbirim/corsproxy:1.2.6
docker run -d --name px4 --network bench -e ALLOW_PRIVATE_NETWORKS=true -e RATE_LIMIT_PER_MINUTE=5 ghcr.io/melihbirim/corsproxy:1.2.6
docker run -d --name cfw --network bench cf-workerd

# load (repeat with -n 40000 -c 200)
docker run --rm --network bench httpd:2.4-alpine ab -k -n 20000 -c 50 'http://px:8080/?url=http://up:9000/'
docker run --rm --network bench httpd:2.4-alpine ab -k -n 20000 -c 50 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/'
docker stats --no-stream px cfw

# security
./probes.sh
```

`ALLOW_PRIVATE_NETWORKS=true` on `px` only exists so corsproxy will talk to an
upstream on a Docker-internal address; `px2` runs with defaults so probes 1
and 2 show what `docker run` gives you with no configuration.
