#!/usr/bin/env bash

docker run --name minio --rm -d -p 9000:9000 \
    -e "MINIO_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
    -e "MINIO_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
    minio/minio server /data
