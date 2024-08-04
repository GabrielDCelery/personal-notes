### Cost allocation tags

Cost allocation tags are an extra feature to get more detailed billing reports within AWS. They either have to be enabled on a per account basis, or at the organizational level in the management account.

When enabled AWS automatically starts adding its own tags, like `aws:createdBy` (which details the identity that created a resource) or `aws:cloudformation:stack-name`.

It is important that these tags `do not get added to the billing reports retroactively(!)`. Also it can take up to `24 hours` for these to take effect and become visible.

After enabling the feature you can also specify custom tags, like `user:something`.

These tags can be used for filtering, so you could tell which costs belong to which environments or departments.


