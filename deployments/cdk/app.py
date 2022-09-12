#!/usr/bin/env python3
import os

import aws_cdk as cdk

from cdk.vpc_stack import VpcStack


app = cdk.App()
vpc_stack = VpcStack(app, "PuppersTestVpcStack")

app.synth()
