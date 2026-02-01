● The DocsExplorer agent has researched AWS CDK documentation and provided comprehensive answers. Here are the key findings:

1. AWS CDK Assets & Asset Manifests

How it works:

- Assets are local files, directories, or Docker images that CDK deploys to AWS
- During cdk deploy, assets are hashed (SHA256), packaged, and uploaded to bootstrap S3/ECR before CloudFormation runs
- Asset manifests (in cdk.out/) contain metadata: source paths, destination buckets/repos, and IAM roles to assume for publishing

Publishing process:

1. Synthesis → generates cloud assembly with asset manifests
2. Asset preparation → archives directories, builds Docker images
3. Asset publishing → uploads to bootstrap S3/ECR using FilePublishingRole/ImagePublishingRole
4. CloudFormation deployment → references published assets by S3 key or ECR URI

5. AWS CDK Permissions

Bootstrap creates 5 IAM roles:

- CloudFormationExecutionRole - Has AdministratorAccess by default, actually creates resources
- DeploymentActionRole - CDK CLI assumes this to orchestrate deployments
- FilePublishingRole - Uploads file/directory assets to S3
- ImagePublishingRole - Pushes Docker images to ECR
- LookupRole - Read-only queries for context lookups (VPCs, subnets, etc.)

Permission chain during deployment:
Your IAM credentials → DeploymentActionRole → FilePublishingRole/ImagePublishingRole →
Pass CloudFormationExecutionRole → CloudFormation assumes it → Creates resources

3. Inline Credentials vs AWS Profiles

Environment Variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY):

- ⚠ CDK blocks these by default - must use --unsafe-creds flag
- Risks: plaintext in bash history, visible to all processes, usually long-term credentials
- Use case: Quick testing only, never production

AWS Profiles (~/.aws/credentials and ~/.aws/config):

- More secure: supports role assumption, SSO, temporary credentials
- Usage: cdk deploy --profile prod-deploy
- Best with SSO for centralized identity management and automatic credential refresh

Recommended priority:

1. AWS SSO (temporary creds, MFA)
2. IAM roles for AWS services (EC2, Lambda, CodeBuild)
3. Profiles with role assumption
4. Credential process tools (aws-vault, 1Password)
5. Static keys only as last resort (rotate every 90 days)

The agent included sources from official AWS CDK documentation for further reading.
