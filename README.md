# puppers
puppers is a mock deployable for deployment testing. It accesses a database and uses a couple environment
variables.
puppers builds two executables: 
  - puppers
  - pupperswebserver

## puppers usage
This creates a puppers user with no login capability and /opt/puppers as its home directory,
then downloads the tarball and unpacks it into /opt/puppers requires some environment variables to function:
 - AWS_REGION : required to find the secret in the current region
 - PUPPERS_SECRET_NAME : puppers connects to a postgres database. This is the name of the database secret

```bash
useradd -m -d /opt/puppers -s /bin/bash puppers
usermod  -L puppers

curl -Lo /opt/puppers_0.0.10_linux_amd64.tar.gz \
https://github.com/natemarks/puppers/releases/download/v0.0.10/puppers_0.0.10_linux_amd64.tar.gz

mkdir -p /opt/puppers

tar -xzvf /opt/puppers_0.0.10_linux_amd64.tar.gz -C /opt/puppers

chown -R puppers:puppers /opt/puppers

runuser -l puppers -c "AWS_REGION=us-east-1 PUPPERS_SECRET_NAME=my_secret nohup /opt/puppers/puppers &"
```


## pupperswebserver usage
This creates a puppers user with no login capability and /opt/puppers as its home directory,
then downloads the tarball and unpacks it into /opt/puppers requires some environment variables to function:
- AWS_REGION : required to find the secret in the current region
- PUPPERS_SECRET_NAME : puppers connects to a postgres database. This is the name of the database secret

```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 709310380790.dkr.ecr.us-east-1.amazonaws.com
docker pull 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:245edeeca8adba53919986eeef5716fdb26579c4
docker run --rm -e AWS_REGION=us-east-1 -e PUPPERS_SECRET_NAME=SecretA720EF05-YA1uGeVL9JKx -p 8080:8080 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:245edeeca8adba53919986eeef5716fdb26579c4
```


### commit test: puppers
Assuming you have an EC2 instance with access to a database and its secret:

Build the executables from the commit and upload them to an S3 bucket that thew EC2 instance can access
```bash
# on the local dev machine
make build
make S3_BUCKET=my_bucket s3_upload
```

On the EC2 instance download and test:
```bash
cd $(mktemp -d)
aws s3 cp s3://my_bucket/puppers/puppers_<commit>.zip .
unzip puppers_<commit>.zip
AWS_REGION=us-east-1 PUPPERS_SECRET_NAME=my_secret ./build/<commit>/liunx/amd64/puppers

```


### commit test: pupperswebserver
On the local dev machine
```bash
make docker-build

```
On the test EC2 instance

```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 709310380790.dkr.ecr.us-east-1.amazonaws.com
docker pull 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:245edeeca8adba53919986eeef5716fdb26579c4
docker run --rm -e AWS_REGION=us-east-1 -e PUPPERS_SECRET_NAME=my_secret -p 8080:8080 709310380790.dkr.ecr.us-east-1.amazonaws.com/puppers:245edeeca8adba53919986eeef5716fdb26579c4
```

## release a new version
On the local dev machine, merge the changes into the main branch and push/pull them.
Bump the version
```bash
make part=patch bump
gpoat
```
In github, draft a new release


The push the release image to  ECR
```bash
make docker-release
```