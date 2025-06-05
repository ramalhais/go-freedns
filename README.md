# go-freedns
This library manages DNS records on freedns.afraid.org

# external-dns-provider-freedns resides here
git clone git@github.com:ramalhais/external-dns-provider-freedns.git

## Multi-arch build
```
git clone https://github.com/ramalhais/external-dns.git

docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
apt install docker-buildx
```
### amd64
```
VER=8

docker build --platform amd64 -t ramalhais/external-dns:freedns-${VER}-amd64 --build-arg ARCH=amd64 .
docker push ramalhais/external-dns:freedns-${VER}-amd64
```

### arm64
```
docker build --platform arm64 -t ramalhais/external-dns:freedns-${VER}-arm64v8 --build-arg ARCH=arm64v8 .
docker push ramalhais/external-dns:freedns-${VER}-arm64v8
```

### manifest
```
docker manifest create \
ramalhais/external-dns:freedns-${VER} \
--amend ramalhais/external-dns:freedns-${VER}-amd64 \
--amend ramalhais/external-dns:freedns-${VER}-arm64v8

docker manifest push ramalhais/external-dns:freedns-${VER}
```
