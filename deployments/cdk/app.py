#!/usr/bin/env python3
"""Deploy puppers test application
Use CDK to deploy puppers for testing
"""

import aws_cdk as cdk

from cdk.vpc_stack import VpcStack
from cdk.rds_stack import RdsStack
from cdk.ec2_stack import Ec2Stack


app = cdk.App()
vpc_stack = VpcStack(app, "PuppersTestVpcStack")
rds_stack = RdsStack(app, "PuppersTestRdsStack", vpc_stack.vpc)
rds_stack.add_dependency(vpc_stack)
ec2_stack = Ec2Stack(
    app,
    "PuppersTestEc2Stack",
    vpc_stack.vpc,
    rds_stack.my_secret,
    rds_stack.instance1,
)
ec2_stack.add_dependency(rds_stack)

app.synth()
