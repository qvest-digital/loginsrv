#!/bin/bash

docker build -t tarent/loginsrv . ;
docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD" ;
docker push tarent/loginsrv ;
