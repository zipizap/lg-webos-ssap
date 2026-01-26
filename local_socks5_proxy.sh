docker rm -f go-socks5-proxy 2>/dev/null || true
echo "*************************************************************"
echo "* SOCKS5 proxy listening on docker-host 127.0.0.1:1080      *"
echo "* Test:   curl --socks5 127.0.0.1:1080 https://ifconfig.me  *"
echo "* STOP with Ctrl-P Ctrl-Q                                   *"
echo "*************************************************************"
docker run \
  -it --rm \
  --name go-socks5-proxy \
  -p 127.0.0.1:1080:1080 \
  -e REQUIRE_AUTH=false \
  serjs/go-socks5-proxy

docker rm -f go-socks5-proxy 2>/dev/null || true



