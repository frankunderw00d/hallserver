#!/bin/bash

vct "$1" "$2"
version=$(cat versionTemporary.txt)
echo "${version}"
rm versionTemporary.txt

CGO_ENABLED=0 go build -ldflags "-s -w" -o hallserver

# 上传${version}版本的镜像后，清除本地中间层镜像，再拉取下来 tag
docker image build -t hallserver:v${version} .
docker image tag hallserver:v${version} 127.0.0.1:5000/hallserver:v${version}
docker image push 127.0.0.1:5000/hallserver:v${version}
docker image rm hallserver:v${version}
docker image rm 127.0.0.1:5000/hallserver:v${version}
docker image prune -f
docker image pull 127.0.0.1:5000/hallserver:v${version}
docker image tag 127.0.0.1:5000/hallserver:v${version} hallserver:v${version}
docker image rm 127.0.0.1:5000/hallserver:v${version}

rm hallserver