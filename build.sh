#!/bin/bash

operator-sdk build  bindong.top/operator/redis-operator:1.0.1
docker push  bindong.top/operator/redis-operator:1.0.1

sed -i "" 's|REPLACE_IMAGE|bindong.top/operator/redis-operator:1.0.1|g' deploy/operator.yaml