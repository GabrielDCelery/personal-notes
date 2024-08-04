### Trusted Advisor

Trusted Advisor is a service that compares your current settings in your environment against AWS's best practices. It is an `account level service` and no need to install anything, `it just works out of the box`.

There are 5 areas that it analyzes:

- Cost
- Performance
- Security
- Fault tolerance
- Service limits

#### What checks do you get with the basic version?

It is important to note that while the service is enabled by default, only a basic version of it runs all the time that runs `7 core checks` on the `basic` and `development` account settings. For any further analysis you cannot enable better tiers separately, for the enhanced features the entire account needs to be moved either to `business` or `enterprise` tiers.

The 7 core checks are:

- S3 bucket permissions (not objects!!)
- unrestricted ports on security groups
- do you have at least one IAM role
- do you have MFA on the root account
- do you have any public EBS snapshots
- do you have any public RDS snapshots
- most common 50 service limits and whether you are about to reach them

#### What checks do you get with the upgraded version?

With the `business` or `enterprise` level you get an additional 115 checks and the `AWS support API`. The latter is critical if you want to use any `Trusted Advisor` functionality programatically. It allows you to:

- you can get the names and IDs of the checks Trusted Advisor performs
- you can request checks automatically
- you can get summaries
- you can obtain the status and results of checks
- open/search support cases or add comments etc...

With the enhanced version you also get `Cloudwatch Integration`. With that you can create automated actions based on what are the results of the Trusted Advisor checks.

#### How to handle issues?

Dependent on the status of your environment Trusted Advisor can tell you if in a certain area:

- Everyting is OK
- Investigation recommended
- Action recommended
