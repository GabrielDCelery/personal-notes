---
title: "Cloudformation import resource"
date: 2025-09-25
tags: ["cloudformation", "aws"]
---

● Yes, you can do this using CloudFormation's resource import feature. Here's how:

Steps to Import an Existing Resource

1. Recreate the Target Group manually with the same physical resource ID (ARN) and settings as the deleted one. If you still have the original ARN, use the same name since target group ARNs include the name.
2. Modify your CloudFormation template to match the manually created resource exactly. The template definition must match the actual resource configuration.
3. Create a change set with import operation:
   aws cloudformation create-change-set \
    --stack-name your-stack-name \
    --change-set-name import-target-group \
    --change-set-type IMPORT \
    --resources-to-import file://resources-to-import.json \
    --template-body file://your-template.yaml
4. Create the resources-to-import.json file:

```json
[
  {
    "ResourceType": "AWS::ElasticLoadBalancingV2::TargetGroup",
    "LogicalResourceId": "MyTargetGroup",
    "ResourceIdentifier": {
      "TargetGroupArn": "arn:aws:elasticloadbalancing:region:account:targetgroup/name/id"
    }
  }
]
```

5. Review and execute the change set:
   aws cloudformation describe-change-set \
    --stack-name your-stack-name \
    --change-set-name import-target-group

aws cloudformation execute-change-set \
 --stack-name your-stack-name \
 --change-set-name import-target-group

Important Notes

- The manually created target group must have the exact same ARN as the deleted one, or you'll need to update any references to it in your stack
- The template definition must exactly match the actual resource configuration (VPC, protocol, port, health check settings, etc.)
- Resource import is only available for https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/resource-import-supported-resources.html, which includes ALB Target Groups
- You cannot modify the resource during import - the template must match reality exactly

This allows you to bring the manually created resource back under CloudFormation management without redeploying the entire stack.
