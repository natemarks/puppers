#!/usr/bin/env python3
# pylint: disable=duplicate-code,too-many-arguments
"""Build the Fargate stack to deploy puppers in ECS EC2 with an NLB
"""
from aws_cdk import (
    Duration,
    Stack,
    aws_autoscaling as autoscaling,
    aws_ec2 as ec2,
    aws_ecs as ecs,
    aws_iam as iam,
    aws_ecr as ecr,
    aws_secretsmanager as sm,
    aws_rds as rds,
    aws_ecs_patterns as ecs_patterns,
)
from constructs import Construct


class Ec2NlbStack(Stack):
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

        asg_instance_role = iam.Role(
            self,
            "InstanceSSM",
            assumed_by=iam.ServicePrincipal("ec2.amazonaws.com"),
        )
        # add the SSM policy so we can manage the instance with SSM
        asg_instance_role.add_managed_policy(
            iam.ManagedPolicy.from_aws_managed_policy_name(
                "AmazonSSMManagedInstanceCore"
            )
        )

        task_role = iam.Role(
            self,
            "Ec2NlbTaskRole",
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
        )
        # permit puppers to access the RDS secret
        task_role.add_to_policy(
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
        task_role.add_to_policy(
            iam.PolicyStatement(
                resources=["*"], actions=["secretsmanager:ListSecrets"]
            )
        )
        cluster = ecs.Cluster(self, "PuppersTestEc2NlbCluster", vpc=target_vpc)
        my_asg = autoscaling.AutoScalingGroup(
            self,
            "PuppersTestEc2NlbASG",
            vpc=target_vpc,
            instance_type=ec2.InstanceType("t3.micro"),
            machine_image=ecs.EcsOptimizedImage.amazon_linux2(),
            desired_capacity=3,
            role=asg_instance_role,
            new_instances_protected_from_scale_in=False,
        )
        capacity_provider = ecs.AsgCapacityProvider(
            self,
            "PuppersTestEc2NlbAsgCapacityProvider",
            auto_scaling_group=my_asg,
        )
        cluster.add_asg_capacity_provider(capacity_provider)
        repo = ecr.Repository.from_repository_name(
            self, "PuppersRepository", "puppers"
        )
        image = ecs.ContainerImage.from_ecr_repository(
            repository=repo, tag="245edeeca8adba53919986eeef5716fdb26579c4"
        )
        ec2nlb = ecs_patterns.NetworkLoadBalancedEc2Service(
            self,
            "MyEc2NlbService",
            cluster=cluster,  # Required
            desired_count=3,  # Default is 1
            task_image_options=ecs_patterns.NetworkLoadBalancedTaskImageOptions(
                image=image,
                container_port=8080,
                environment={
                    "PUPPERS_SECRET_NAME": secret.secret_name,
                    "AWS_REGION": "us-east-1",
                },
                task_role=task_role,
            ),
            memory_limit_mib=512,  # Default is 512
        )  # Default is True
        ec2nlb.service.connections.allow_to_default_port(rds_instance)
        my_asg.connections.allow_to_default_port(rds_instance)
        my_asg.connections.allow_from_any_ipv4(
            port_range=ec2.Port.tcp_range(32768, 65535),
            description="allow incoming traffic from ALB",
        )
        ec2nlb.target_group.configure_health_check(
            enabled=True,
            # path="/heartbeat",
            interval=Duration.seconds(30),
            # timeout=Duration.seconds(2),
            port="traffic-port",
        )
