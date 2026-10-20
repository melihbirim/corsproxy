#!/bin/sh
# Seven security probes against corsproxy (px, px2 defaults, px3 origin
# allowlist, px4 rate limit) and Cloudflare's example Worker (cfw), all on
# the "bench" Docker network. See README.md for how to start them.
C='docker run --rm --network bench curlimages/curl:8.10.1 -s -m 8'
echo "1 private-network target, default config"
$C -o /dev/null -w 'corsproxy %{http_code}\n' 'http://px2:8080/?url=http://up:9000/'
$C -o /dev/null -w 'worker    %{http_code}\n' 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/'
echo "2 upstream 302 -> 169.254.169.254"
$C -o /dev/null -w 'corsproxy %{http_code}\n' 'http://px2:8080/?url=http://up:9000/redir'
$C -o /dev/null -w 'worker    %{http_code} %{redirect_url}\n' 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/redir'
echo "3 Authorization + Cookie forwarded to target?"
printf 'corsproxy '; $C -H 'Authorization: Bearer SECRET' -H 'Cookie: s=SECRET' 'http://px:8080/?url=http://up:9000/echo' | grep -oE 'Authorization|Cookie' | tr '\n' ' '; echo
printf 'worker    '; $C -H 'Authorization: Bearer SECRET' -H 'Cookie: s=SECRET' 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/echo' | grep -oE 'Authorization|Cookie' | tr '\n' ' '; echo
echo "4 Origin: https://evil.example"
$C -D - -o /dev/null -H 'Origin: https://evil.example' 'http://px3:8080/?url=http://up:9000/' | grep -i 'allow-origin' | sed 's/^/corsproxy /'
$C -D - -o /dev/null -H 'Origin: https://evil.example' 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/' | grep -i 'allow-origin' | sed 's/^/worker    /'
echo "5 DELETE"
$C -o /dev/null -w 'corsproxy %{http_code}\n' -X DELETE 'http://px:8080/?url=http://up:9000/echo'
$C -o /dev/null -w 'worker    %{http_code}\n' -X DELETE 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/echo'
echo "6 20 MB upstream body"
$C -o /dev/null -w 'corsproxy %{http_code} %{size_download} bytes\n' 'http://px:8080/?url=http://up:9000/big?mb=20'
$C -o /dev/null -w 'worker    %{http_code} %{size_download} bytes\n' 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/big?mb=20'
echo "7 rate limit 5/min, 7 requests"
printf 'corsproxy '; for i in 1 2 3 4 5 6 7; do $C -o /dev/null -w '%{http_code} ' 'http://px4:8080/?url=http://up:9000/'; done; echo
printf 'worker    '; for i in 1 2 3 4 5 6 7; do $C -o /dev/null -w '%{http_code} ' 'http://cfw:8787/corsproxy/?apiurl=http://up:9000/'; done; echo
