#!/usr/bin/env python3
"""Build the VPC stack for puppers tst
"""
from aws_cdk import (
    Stack,
    aws_ec2 as ec2,
)
from constructs import Construct


class VpcStack(Stack):
    """VPC stack subclass"""

    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        self.vpc = ec2.Vpc(
            self,
            "PuppersTestVpc",
            cidr="10.44.0.0/16",
            subnet_configuration=[
                ec2.SubnetConfiguration(
                    name="pupperstest_public",
                    subnet_type=ec2.SubnetType.PUBLIC,
                ),
                ec2.SubnetConfiguration(
                    name="pupperstest_private",
                    subnet_type=ec2.SubnetType.PRIVATE_WITH_NAT,
                ),
                ec2.SubnetConfiguration(
                    name="pupperstest_isolated",
                    subnet_type=ec2.SubnetType.PRIVATE_ISOLATED,
                ),
            ],
        )

        self.vpc.add_interface_endpoint(
            "EC2", service=ec2.InterfaceVpcEndpointAwsService.EC2
        )
        self.vpc.add_interface_endpoint(
            "EC2_MESSAGES",
            service=ec2.InterfaceVpcEndpointAwsService.EC2_MESSAGES,
        )
        self.vpc.add_interface_endpoint(
            "SSM", service=ec2.InterfaceVpcEndpointAwsService.SSM
        )
        self.vpc.add_interface_endpoint(
            "SSM_MESSAGES",
            service=ec2.InterfaceVpcEndpointAwsService.SSM_MESSAGES,
        )
