# go-freedns
This library manages DNS records on freedns.afraid.org

## Multi-arch build

docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
apt install docker-buildx

### amd64
```
docker build --platform amd64 -t ramalhais/external-dns:freedns-7-amd64 --build-arg ARCH=amd64 .
docker push ramalhais/external-dns:freedns-7-amd64
```

### arm64
```
docker build --platform arm64 -t yramalhais/external-dns:freedns-7-arm64v8 --build-arg ARCH=arm64v8 .
docker push ramalhais/external-dns:freedns-7-arm64v8
```

### manifest
```
docker manifest create \
ramalhais/external-dns:freedns-7 \
--amend ramalhais/external-dns:freedns-7-amd64 \
--amend ramalhais/external-dns:freedns-7-arm64v8

docker manifest push ramalhais/external-dns:freedns-7
```
