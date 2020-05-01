#!/bin/bash
set -x
awslocal s3api create-bucket \
        --bucket converter-service-source \
        --acl public-read-write >/dev/null
echo "converter-service-source created"

