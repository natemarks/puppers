#!/usr/bin/env python3
"""Deploy puppers test application
Use CDK to deploy puppers for testing
"""

import aws_cdk as cdk

from cdk.vpc_stack import VpcStack


app = cdk.App()
vpc_stack = VpcStack(app, "PuppersTestVpcStack")

app.synth()
