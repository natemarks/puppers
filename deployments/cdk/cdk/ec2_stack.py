#!/usr/bin/env python3
# pylint: disable=line-too-long,too-many-arguments,
"""Build the EC2 stack to deploy puppers directly on an ec2 instance
"""
from aws_cdk import (
    Stack,
    aws_ec2 as ec2,
    aws_ecs as ecs,
    aws_iam as iam,
    aws_secretsmanager as sm,
    aws_rds as rds,
)
from constructs import Construct


class Ec2Stack(Stack):
    """VPC stack subclass"""

    def __init__(
        self,
        scope: Construct,
        construct_id: str,
        target_vpc,
        secret: sm.Secret,
        rds_instance: rds.DatabaseInstance,
        **kwargs
    ) -> None:
        super().__init__(scope, construct_id, **kwargs)

        with open("./userdata/puppers.sh", encoding="utf-8") as file:
            user_data = file.read()

        ec2_optimized_ami = ecs.EcsOptimizedImage.amazon_linux2()
        # create role for EC2 Instancs
        role = iam.Role(
            self,
            "InstanceSSM",
            assumed_by=iam.ServicePrincipal("ec2.amazonaws.com"),
        )
        # add the SSM policy so we can manage the instance with SSM
        role.add_managed_policy(
            iam.ManagedPolicy.from_aws_managed_policy_name(
                "AmazonSSMManagedInstanceCore"
            )
        )
        # Permit the instance to write to cloudwatch. Required for puppers
        role.add_managed_policy(
            iam.ManagedPolicy.from_aws_managed_policy_name(
                "CloudWatchAgentServerPolicy"
            )
        )
        # Permit the instance to access docker repository Required for puppers
        role.add_managed_policy(
            iam.ManagedPolicy.from_aws_managed_policy_name(
                "AmazonEC2ContainerRegistryReadOnly"
            )
        )
        # permit puppers to access the RDS secret
        role.add_to_policy(
            iam.PolicyStatement(
                resources=[secret.secret_full_arn],
                actions=[
                    "secretsmanager:GetResourcePolicy",
                    "secretsmanager:GetSecretValue",
                    "secretsmanager:DescribeSecret",
                    "secretsmanager:ListSecretVersionIds",
                ],
            )
        )
        role.add_to_policy(
            iam.PolicyStatement(resources=["*"], actions=["secretsmanager:ListSecrets"])
        )
        # permit EC2 instance to download the commit builds of puppers from S3
        role.add_to_policy(
            iam.PolicyStatement(resources=["*"], actions=["s3:ListAllMyBuckets"])
        )
        role.add_to_policy(
            iam.PolicyStatement(
                resources=[
                    "arn:aws:s3:::com.imprivata.709310380790.us-east-1"
                    ".devops-artifacts",
                    "arn:aws:s3:::com.imprivata.709310380790.us-east-1"
                    ".devops-artifacts/*",
                ],
                actions=["s3:ListBucket"],
            )
        )
        role.add_to_policy(
            iam.PolicyStatement(
                resources=[
                    "arn:aws:s3:::com.imprivata.709310380790.us-east-1"
                    ".devops-artifacts/puppers/*"
                ],
                actions=["s3:GetObject"],
            )
        )
        self.instance = ec2.Instance(
            self,
            "PupperstestInstance",
            instance_type=ec2.InstanceType("t3.nano"),
            machine_image=ec2_optimized_ami,
            vpc=target_vpc,
            vpc_subnets=ec2.SubnetSelection(
                subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS
            ),
            role=role,
            user_data=ec2.UserData.custom(user_data),
        )
        self.instance.connections.allow_to_default_port(rds_instance)
