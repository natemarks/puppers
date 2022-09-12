#!/usr/bin/env bash
set -Eeuo pipefail

yum install -y unzip
curl https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip -o awscliv2.zip
unzip awscliv2.zip
sudo ./aws/install

useradd -m -d /opt/puppers -s /bin/bash puppers
usermod  -L puppers

touch /tmp/puppers.txt
curl -Lo /opt/puppers_0.0.10_linux_amd64.tar.gz \
https://github.com/natemarks/puppers/releases/download/v0.0.10/puppers_0.0.10_linux_amd64.tar.gz

mkdir -p /opt/puppers

tar -xzvf /opt/puppers_0.0.10_linux_amd64.tar.gz -C /opt/puppers

chown -R puppers:puppers /opt/puppers

runuser -l puppers -c "AWS_REGION=us-east-1 PUPPERS_SECRET_NAME=SecretA720EF05-YA1uGeVL9JKx nohup /opt/puppers/puppers &"
