#!/usr/bin/env python3
# pylint: disable=duplicate-code,too-many-arguments
"""Build the Fargate stack to deploy puppers in ECS fargate with an ALB
"""
from aws_cdk import (
    Duration,
    Stack,
    aws_ecs as ecs,
    aws_iam as iam,
    aws_ecr as ecr,
    aws_secretsmanager as sm,
    aws_rds as rds,
    aws_ecs_patterns as ecs_patterns,
)
from constructs import Construct


class FargateStack(Stack):
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
        task_role = iam.Role(
            self,
            "FargateTaskRole",
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
            iam.PolicyStatement(resources=["*"], actions=["secretsmanager:ListSecrets"])
        )
        cluster = ecs.Cluster(self, "PuppersTestFargateCluster", vpc=target_vpc)
        repo = ecr.Repository.from_repository_name(self, "PuppersRepository", "puppers")
        image = ecs.ContainerImage.from_ecr_repository(
            repository=repo, tag="245edeeca8adba53919986eeef5716fdb26579c4"
        )
        fgs = ecs_patterns.ApplicationLoadBalancedFargateService(
            self,
            "MyFargateService",
            cluster=cluster,  # Required
            cpu=512,  # Default is 256
            desired_count=3,  # Default is 1
            task_image_options=ecs_patterns.ApplicationLoadBalancedTaskImageOptions(
                image=image,
                container_port=8080,
                environment={"PUPPERS_SECRET_NAME": secret.secret_name},
                task_role=task_role,
            ),
            memory_limit_mib=2048,  # Default is 512
            public_load_balancer=True,
        )  # Default is True
        fgs.service.connections.allow_to_default_port(rds_instance)
        fgs.target_group.configure_health_check(
            enabled=True,
            path="/heartbeat",
            interval=Duration.seconds(120),
            timeout=Duration.seconds(2),
            port="8080",
        )
