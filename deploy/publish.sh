#!/bin/sh


yc config profile create temp-sa
yc config set service-account-key ./auth_key.json

echo $(yc iam create-token) | docker login \
  --username iam \
  --password-stdin \
  cr.yandex

docker build -f ./Dockerfile -t chartdb_back:latest ..
docker tag chartdb_back:latest cr.yandex/crpmjlujd7cnae91rici/chartdb_back:latest
docker push cr.yandex/crpmjlujd7cnae91rici/chartdb_back:latest

docker logout cr.yandex

yc config profile activate default
yc config profile delete temp-sa
