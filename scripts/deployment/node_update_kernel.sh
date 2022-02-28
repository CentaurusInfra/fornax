#!/bin/bash

set -x

cd ~

git clone -b dev-next-fornax https://github.com/CentaurusInfra/mizar.git

cd mizar/

./kernelupdate.sh

