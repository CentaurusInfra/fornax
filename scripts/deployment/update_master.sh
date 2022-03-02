#!/bin/bash
PYTHON_VERSION=${PYTHON_VERSION:-"3.7.12"}

sudo apt-get remove python3.6
sudo apt-get remove --auto-remove python3.6
sudo apt-get purge python3.6
sudo apt-get purge --auto-remove python3.6

# Install python 3.7.12 since apt install always install python3.7.5
sudo apt update -y
sudo apt install -y build-essential zlib1g-dev libncurses5-dev libgdbm-dev libnss3-dev libssl-dev libsqlite3-dev libreadline-dev libffi-dev wget libbz2-dev

sudo apt-get remove -y --purge python3.7

cd /tmp; 
wget https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz
tar -xf Python-${PYTHON_VERSION}.tgz
rm -rf Python-${PYTHON_VERSION}.tgz
cd Python-${PYTHON_VERSION}
./configure --enable-optimizations
sudo make install
python3 --version

ln -vfns  /tmp/Python-3.7.12/python /usr/local/bin/python3


#!/bin/bash

# Install go and related
cd /tmp; wget https://dl.google.com/go/go1.13.9.linux-amd64.tar.gz
rm -rf /usr/local/go && rm -rf /usr/local/bin/go && tar -C /usr/local -xzf go1.13.9.linux-amd64.tar.gz
rm -rf go1.13.9.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
sudo apt-get install -y protobuf-compiler libprotobuf-dev
GO111MODULE="on" go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
GO111MODULE="on" go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
GO111MODULE="on" go get github.com/smartystreets/goconvey@v1.6.4

apt-get install -y build-essential clang-7 llvm-7 libelf-dev python3.8 python3-pip libcmocka-dev lcov python3.8-dev python3-apt pkg-config
python3 -m pip install --user grpcio-tools
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin:/usr/local/bin
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN
