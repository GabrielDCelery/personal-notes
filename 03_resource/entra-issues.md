> passwordless authentication via the authenticator app - where can I see how it is set up on azure

⏺ In Azure (Microsoft Entra ID), passwordless authentication settings are configured in a few places:

Primary Location - Authentication Methods

1. Azure Portal → Microsoft Entra ID (formerly Azure AD)
2. Navigate to: Security → Authentication methods → Policies
3. Here you'll find:
   - Microsoft Authenticator - Shows if it's enabled and for which users/groups
   - Click on "Microsoft Authenticator" to see detailed configuration:
     - Target users/groups
     - Authentication mode (Any, or Passwordless only)
     - Feature settings (number matching, application name display, etc.)

Additional Related Settings

MFA Configuration:

- Microsoft Entra ID → Security → Multifactor authentication
- Shows traditional MFA settings

Conditional Access:

- Microsoft Entra ID → Security → Conditional Access → Policies
- Shows policies that may require passwordless authentication

Sign-in Logs:

- Microsoft Entra ID → Monitoring → Sign-in logs
- Filter by authentication method to see actual usage

Per-User Status:

- Microsoft Entra ID → Users → select a user → Authentication methods
- Shows what methods each user has registered (including Authenticator app)

The main configuration you're looking for is under Security → Authentication methods → Microsoft Authenticator, which controls whether passwordless authentication is enabled and for whom.

> where should i be able to see microsoft emails which website to visit

⏺ You can access Microsoft emails through these websites:

Primary Email Portals

Outlook Web App (Most Common):

- https://outlook.office.com - For Microsoft 365/Office 365 accounts (work/school)
- https://outlook.live.com - For personal Microsoft accounts (Hotmail, Outlook.com)

Alternative URLs

Unified Portal:

- https://outlook.com - Redirects to the appropriate login based on your account

Microsoft 365 Portal:

- https://office.com → Click on Outlook icon after signing in

For Work/Company Accounts

If your company uses custom domains, you might also be able to access through:

- https://mail.yourdomain.com (if configured by your IT)
- Or through your company's specific Microsoft 365 tenant URL

Mobile Apps

- Outlook mobile app (iOS/Android) - Search "Microsoft Outlook" in app stores

---

For your company email with passwordless authentication, you'll most likely use:

- https://outlook.office.com
- Sign in with your work email address
- When prompted, approve the sign-in notification in your Microsoft Authenticator app (passwordless)

> if I wanted to see the exchange server of my companys account that is on it@mycompany.com and I am using my own email my@mycompany.com how do I switch

⏺ To switch between multiple email accounts in the same company Exchange/Outlook environment, you have several options:

Option 1: Add Additional Account in Outlook Web (Quickest)

In Outlook Web App (https://outlook.office.com):

1. Click your profile picture (top right)
2. Click Open another mailbox
3. Type: it@mycompany.com
4. Click Open

This opens the other mailbox in a new tab (requires you have access permissions).

Option 2: Switch Accounts Completely

Sign out and sign in as different user:

1. Click your profile picture (top right)
2. Click Sign out
3. Go to https://outlook.office.com
4. Sign in with: it@mycompany.com
5. Authenticate via your Authenticator app

Option 3: Use Multiple Browser Profiles/Windows

- Private/Incognito window: Open incognito, sign in with it@mycompany.com
- Different browser: Use Chrome for one account, Edge for another
- Browser profiles: Create separate Chrome/Edge profiles for each account

Option 4: Desktop Outlook (If installed)

Add multiple accounts:

1. File → Account Settings → Account Settings
2. Click New → Add it@mycompany.com
3. Switch between accounts using the account selector dropdown

Check Your Access First

You need to have one of these permissions to access it@mycompany.com:

- Delegate access - IT gave you permission
- Shared mailbox access - It's configured as a shared mailbox
- Login credentials - You have the password/authentication rights

If you don't have access, contact your IT admin to grant permissions to the it@mycompany.com mailbox.

> where can I see the mailbox and see if it got shared access

⏺ You can check if you have shared access to the it@mycompany.com mailbox in several places:

Option 1: Check in Outlook Web App (Easiest)

https://outlook.office.com (signed in as my@mycompany.com):

1. Look in the left sidebar under your main mailbox
2. Shared mailboxes you have access to typically appear automatically as additional folders
3. Or try: Click your profile picture → Open another mailbox → Type "it@mycompany.com"
   - If it opens, you have access
   - If denied, you don't have permissions

Option 2: Microsoft 365 Admin Center (Admin Only)

https://admin.microsoft.com (requires admin permissions):

1. Navigate to: Users → Active users
2. Search for and click on it@mycompany.com
3. Go to the Mail tab
4. Look for Mailbox permissions or Mailbox delegation
5. Check if your account (my@mycompany.com) is listed under:
   - Full Access
   - Send As
   - Send on Behalf

Option 3: Exchange Admin Center

https://admin.exchange.microsoft.com:

1. Go to Recipients → Mailboxes
2. Search for and click it@mycompany.com
3. Click on Delegation tab
4. Check these sections:
   - Full Access - Can you access the entire mailbox?
   - Send As - Can you send as this mailbox?
   - Send on Behalf - Can you send on behalf?

Option 4: Check Your Own Account Settings

https://admin.microsoft.com → Users → Search for my@mycompany.com → Mail tab:

- Sometimes shows mailboxes you have access to

Quick Test Method

Simply try to access it:

- Go to https://outlook.office.com/mail/it@mycompany.com
- If it loads, you have access
- If error/denied, you don't

---

Note: You need to be an admin or have the IT team check this for you if you can't access the admin centers listed above.

> what is the url to microsoft exchange admin server

⏺ Microsoft Exchange Admin Center URLs

For Microsoft 365 / Exchange Online (Cloud)

New Exchange Admin Center (Current):

- https://admin.exchange.microsoft.com

Classic Exchange Admin Center (Legacy):

- https://outlook.office365.com/ecp

Alternative Access:

- Go to https://admin.microsoft.com (Microsoft 365 Admin Center)
- Click Show all in left sidebar → Admin centers → Exchange

---

For On-Premises Exchange Server

If your company runs Exchange on their own servers, the URL will be specific to your organization, typically:

- https://mail.mycompany.com/ecp
- https://exchange.mycompany.com/ecp
- https://webmail.mycompany.com/ecp

(Replace "mycompany.com" with your actual domain)

To find your specific URL:

1. Ask your IT department
2. Or try adding /ecp to your Outlook Web Access URL

---

Requirements

- You need admin permissions (Exchange Administrator role or higher)
- For cloud: Must be a Global Admin or Exchange Administrator
- If you don't have access, you'll see a permissions error

For your company: Since you're using Microsoft Authenticator for passwordless auth, you're likely using Microsoft 365, so start with https://admin.exchange.microsoft.com

> if my org set up passeordless sign in request using the microsoft authenticator how can I restore that if I switched phones

⏺ Restoring Passwordless Authentication After Phone Switch

You have several options to restore Microsoft Authenticator on your new phone:

Option 1: Cloud Backup & Restore (Easiest)

If you enabled backup on your old phone:

On your NEW phone:

1. Install Microsoft Authenticator app
2. Sign in with your personal Microsoft account (the one you used for backup)
3. The app will ask to restore from backup
4. Tap Recover or Restore from backup
5. Your work account should restore automatically

Backup requirements:

- iPhone: iCloud backup was enabled
- Android: Google account backup was enabled

---

Option 2: Manual Re-registration (Most Common)

If you didn't have backup enabled, you'll need to re-register:

A. If You Still Have Access to Your Account

1. Go to https://mysignins.microsoft.com/security-info
2. Sign in with my@mycompany.com (using password or another method)
3. Find Microsoft Authenticator in the list
4. Click Delete or Remove on the old registration
5. Click Add method → Authenticator app
6. Follow prompts to scan QR code with your new phone

B. If You Can't Sign In (No Other Auth Method)

Contact your IT/Help Desk immediately:

- They can verify your identity
- They'll reset your authentication methods
- You'll then re-register the Authenticator app on your new phone

---

Option 3: Use Alternative Sign-in Method

If your org set up backup authentication:

- Use SMS/phone call if configured
- Use email verification if available
- Use security questions
- Use FIDO2 security key if you have one
- Use backup codes (if you saved them)

To check/use alternatives:

1. On the sign-in screen, click Sign in another way or Use a different method
2. Select available alternative
3. Once signed in, update Authenticator app

---

Option 4: Account Recovery via Company Policy

Some organizations have self-service options:

Go to https://passwordreset.microsoftonline.com

- May allow you to verify identity via alternate methods
- Can reset/update security info

---

Prevention for Next Time

Enable cloud backup NOW on your new phone:

In Microsoft Authenticator:

1. Tap ☰ (menu) → Settings
2. Enable Cloud backup
3. Sign in with your personal Microsoft account
4. This will backup your accounts for future phone switches

---

Immediate Action

If you're locked out right now:

1. Contact your IT Help Desk - they can reset your auth methods
2. Provide proof of identity (employee ID, manager verification, etc.)
3. They'll enable you to re-register Authenticator on your new phone

If you can still sign in:

1. Go to https://mysignins.microsoft.com/security-info
2. Update your Authenticator app registration to your new phone

> if my company disabled mfa via phone number and only allows connection via authenticated app and passwordless sign in how to set up one of our workers who is a user in our organization

⏺ Setting Up a New User with Passwordless-Only Authentication

When your org requires Authenticator app but disabled phone MFA, you have a "chicken and egg" problem - the user needs to sign in to set up the app, but needs the app to sign in. Here are the solutions:

Option 1: Temporary Access Pass (TAP) - Recommended

Microsoft's modern solution for this exact scenario:

As an Admin:

1. Go to https://entra.microsoft.com (or https://portal.azure.com → Microsoft Entra ID)
2. Navigate to: Users → All users
3. Find and click the new user's account
4. In the left menu, click Authentication methods
5. Click + Add authentication method
6. Select Temporary Access Pass
7. Configure:
   - Lifetime: 1 hour to 24 hours (choose appropriate time)
   - One-time use: Yes (recommended) or No

8. Click Add
9. Copy the generated pass and securely send it to the user

As the New User:

1. Go to https://outlook.office.com or company sign-in page
2. Enter email: newuser@mycompany.com
3. When prompted for authentication, enter the Temporary Access Pass
4. You'll be prompted to set up Microsoft Authenticator:
   - Install Microsoft Authenticator app on your phone
   - Scan the QR code shown on screen
   - Approve the test notification

5. Complete the setup
6. The TAP expires/becomes invalid after use or timeout

---

Option 2: Admin Pre-Registration (Alternative)

Admin sets up the user's authentication methods first:

1. https://entra.microsoft.com → Users → Select user
2. Authentication methods → Require re-register MFA
3. Or use a staged rollout approach

---

Option 3: Initial Password Sign-In (If Enabled)

If your org allows one-time password sign-in:

1. Admin creates user with temporary password
2. User signs in with password on first login
3. Immediately prompted to set up Authenticator app
4. After setup, password sign-in is disabled automatically

Check if this is enabled:

- https://entra.microsoft.com → Security → Authentication methods → Policies
- Check if password is still an allowed method for registration

---

Option 4: Admin-Assisted Registration Session

In-person or via screen share:

1. Admin helps user during initial setup
2. User installs Authenticator app on their phone first
3. Admin provides credentials or TAP
4. Guide user through registration process in real-time
5. Verify it works before ending session

---

Recommended Setup Process (Step-by-Step)

Admin Tasks:

1. Create the user account (if not already created)
   - https://admin.microsoft.com → Users → Active users → Add a user

2. Generate Temporary Access Pass:
   - https://entra.microsoft.com → Users → Find user → Authentication methods
   - Add authentication method → Temporary Access Pass
   - Set lifetime: 8-24 hours (enough time for user to complete setup)
   - One-time use: Yes
   - Copy the TAP

3. Send to user securely:
   - Email the TAP to personal email or
   - Communicate via Teams/Slack if accessible or
   - Provide in person

4. Provide instructions (send them the user steps below)

User Tasks:

1. Install Microsoft Authenticator on your phone (before attempting to sign in)
   - iOS: App Store
   - Android: Google Play Store

2. Sign in to Microsoft 365:
   - Go to https://outlook.office.com
   - Enter: yourname@mycompany.com
   - When prompted, enter the Temporary Access Pass provided by IT

3. Set up Authenticator app:
   - You'll see "More information required" screen
   - Click Next
   - Select Mobile app or Authenticator app
   - Choose: Use for verification or passwordless sign-in
   - Scan the QR code with your Authenticator app
   - Approve the test notification
   - Click Done

4. Enable passwordless (if not automatic):
   - In Authenticator app, tap on your work account
   - Tap Enable phone sign-in
   - Follow prompts to complete setup

5. Test it:
   - Sign out
   - Sign back in
   - Should now use passwordless authentication

---

Verify TAP is Available in Your Tenant

Check if TAP is enabled:

1. https://entra.microsoft.com → Security → Authentication methods → Policies
2. Look for Temporary Access Pass
3. Ensure it's Enabled for target users/groups
4. If not enabled, click to enable it

---

Troubleshooting

If TAP option is not available:

- Your tenant might need license (usually included in Microsoft 365)
- May need Global Admin to enable it first
- Alternative: Use password-based registration if still allowed

Best Practice:

- Set TAP lifetime to minimum needed (1-8 hours)
- Use one-time TAPs
- Monitor usage via sign-in logs
- Have a clear onboarding process documented

---

Temporary Access Pass is the modern, secure solution Microsoft designed specifically for this passwordless onboarding scenario.

> how to view what roles an entra group has

⏺ How to View Roles Assigned to an Entra ID Group

Option 1: Check from the Group (Recommended)

https://entra.microsoft.com (or https://portal.azure.com → Microsoft Entra ID):

1. Navigate to: Groups → All groups
2. Search for and click on the group
3. In the left menu, look for Assigned roles (under "Manage")
4. This shows all Entra ID roles assigned to this group

Shows:

- Role name
- Description
- Assignment type (Eligible vs Active if using PIM)

---

Option 2: Check from Roles Perspective

https://entra.microsoft.com:

1. Go to Roles & admins → Roles & admins (or Identity → Roles & administrators)
2. Click on a specific role (e.g., "Global Administrator", "User Administrator")
3. Click Assignments
4. Look through the list to see if your group is assigned
5. You can use the search/filter to find the specific group

---

Option 3: Check Role Assignments Overview

To see all role assignments at once:

1. https://entra.microsoft.com → Roles & admins → Roles & admins
2. Browse through roles and click each to see assignments
3. Or use Privileged Identity Management (PIM) if enabled:
   - Identity Governance → Privileged Identity Management → Microsoft Entra roles
   - Click Assignments to see all active/eligible assignments

---

Option 4: Check Azure RBAC Roles (Different from Entra Roles)

For Azure resource roles (subscription/resource level):

1. Go to https://portal.azure.com
2. Navigate to your Subscription or Resource Group
3. Click Access control (IAM)
4. Click Role assignments tab
5. Search for the group name
6. Shows roles like "Owner", "Contributor", "Reader", etc.

---

Important Distinction

Entra ID Roles (formerly Azure AD roles):

- Control access to Entra ID resources (users, groups, applications)
- Examples: Global Administrator, User Administrator, Groups Administrator
- View at: https://entra.microsoft.com → Groups → [Group] → Assigned roles

Azure RBAC Roles:

- Control access to Azure resources (VMs, databases, storage)
- Examples: Owner, Contributor, Reader
- View at: Resource → Access control (IAM)

---

Using PowerShell (Alternative)

To check Entra ID roles assigned to a group:

# Connect

Connect-MgGraph -Scopes "RoleManagement.Read.Directory"

# Get group ID

$group = Get-MgGroup -Filter "displayName eq 'YourGroupName'"

# Get role assignments for the group

Get-MgRoleManagementDirectoryRoleAssignment -Filter "principalId eq '$($group.Id)'" | ForEach-Object {
$roleId = $\_.RoleDefinitionId
$role = Get-MgRoleManagementDirectoryRoleDefinition -UnifiedRoleDefinitionId $roleId
$role.DisplayName
}

---

Quick Summary

Fastest method:

1. Go to https://entra.microsoft.com
2. Groups → All groups → Search for your group → Click it
3. Click Assigned roles in left menu
4. See all Entra ID roles assigned to that group

This shows you exactly what administrative permissions that group has in your tenant.

> if my frontend was using a jwt token from an azure entra login where would I find what populates the jwt token "custom:AppRole"

⏺ Where to Find JWT Token Claims Configuration in Azure Entra ID

The custom:AppRole claim is likely configured through App Roles and Token Configuration in your Azure Entra app registration. Here's where to check:

---

Option 1: App Roles Definition (Most Likely)

https://entra.microsoft.com or https://portal.azure.com:

1. Microsoft Entra ID → App registrations
2. Find and click your application (the one your frontend uses)
3. In the left menu, click App roles
4. Here you'll see defined roles like:
   - Admin
   - User
   - Editor
   - etc.

This defines what roles exist, but doesn't assign them to users yet.

---

Option 2: Token Configuration (Custom Claims)

In the same App Registration:

1. App registrations → Your app → Token configuration
2. Check for optional claims configured here
3. Look for custom claim mappings that might create custom:AppRole

To add/modify:

- Click Add optional claim
- Or check Add groups claim if roles come from groups
- Custom claims can be added here

---

Option 3: Enterprise Application (Role Assignments)

Where users/groups are actually assigned the roles:

https://entra.microsoft.com:

1. Enterprise applications → All applications
2. Find and click your application (same name as app registration)
3. Click Users and groups in left menu
4. This shows which users/groups have which roles assigned
5. Each user/group will have a role like "Admin", "User", etc.

This is where the actual assignment happens - defining which users get which app roles.

---

Option 4: Claims Mapping Policy (Advanced)

For custom claim transformations:

Via PowerShell (cannot be done in portal):

Connect-MgGraph -Scopes "Policy.Read.All", "Application.Read.All"

# Get your app

$app = Get-MgServicePrincipal -Filter "displayName eq 'YourAppName'"

# Check if claims mapping policy is attached

Get-MgServicePrincipalClaimMappingPolicy -ServicePrincipalId $app.Id

This is used for advanced scenarios where you need to transform standard claims into custom formats.

---

Option 5: Manifest (Raw Configuration)

Check the app manifest directly:

1. App registrations → Your app → Manifest
2. Look for appRoles section in the JSON:

"appRoles": [
{
"allowedMemberTypes": ["User"],
"description": "Administrators have full access",
"displayName": "Admin",
"id": "some-guid",
"isEnabled": true,
"value": "Admin"
},
{
"allowedMemberTypes": ["User"],
"description": "Regular users",
"displayName": "User",
"id": "another-guid",
"isEnabled": true,
"value": "User"
}
]

The value field is what appears in the token.

---

How to Verify What's Actually in Your Token

Decode your actual JWT token:

1. Get the token from your browser (DevTools → Application → Storage)
2. Go to https://jwt.ms or https://jwt.io
3. Paste your token
4. Look for the custom:AppRole claim in the decoded payload

Standard Azure Entra claims for roles:

- roles - array of app roles
- groups - array of group IDs (if configured)
- Custom claims you configured

---

Important Note About "custom:AppRole"

The custom:AppRole format with the custom: prefix is unusual for Azure Entra ID.

Typically, Azure Entra uses:

- roles claim (array): ["Admin", "User"]
- Not custom:AppRole

Possible scenarios:

1. Claims Mapping Policy Applied

Someone configured a custom claims mapping to transform roles → custom:AppRole

2. Custom Optional Claim

Configured in Token Configuration with extension attributes

3. Different Identity Provider

Are you sure it's Azure Entra? The custom: prefix is common in AWS Cognito, not Azure AD.

---

Recommended Steps

1. Decode your actual token at https://jwt.ms to see all claims
2. Check App Roles: App registrations → Your app → App roles
3. Check Token Configuration: App registrations → Your app → Token configuration
4. Check Role Assignments: Enterprise applications → Your app → Users and groups
5. Check the Manifest: Look for appRoles and optionalClaims sections

---

If You Need to Find the Specific App

If you don't know which app registration:

1. Look at your frontend code for the client ID or application ID
2. In Azure: App registrations → All applications
3. Search by the application/client ID
4. Or check your frontend config files for values like:
   - clientId
   - applicationId
   - tenantId

Once you find the right app, check the sections above to see how custom:AppRole is populated.

> if i got an entra group and and clicked on it in the applications section I see multiple applications that are directly assigned - where are these created and where can they be viewed

⏺ Where Applications Shown in Group Assignments Are Created and Viewed

When you see applications in a group's "Applications" section, these are Enterprise Applications that the group has been assigned to. Here's where to find and manage them:

---

Where to View the Applications

Option 1: Enterprise Applications (Main Location)

https://entra.microsoft.com:

1. Identity → Applications → Enterprise applications → All applications
2. Search for the application name you saw in the group
3. Click on it to see details
4. Go to Users and groups to see all group/user assignments
   - You'll see your group listed here with the assigned role

This is where the application exists as a deployed/usable app in your tenant.

---

Option 2: App Registrations (Where Created)

https://entra.microsoft.com:

1. Identity → Applications → App registrations → All applications
2. Search for the same application name
3. This is where the app was originally created/registered

Two types:

- Apps created in your tenant - Shows in "Owned applications"
- Third-party apps - May not show here (only as Enterprise App)

---

Understanding the Relationship

App Registration → Creates the app definition (code, permissions, settings)
↓
Enterprise Application → The instance of that app in your tenant (assignments, SSO settings)
↓
Group Assignment → Groups/users who can access the Enterprise Application

---

How These Applications Are Created

Scenario 1: Internal Apps (Created by Your Org)

Created in App Registrations:

1. App registrations → New registration
2. Developer registers an app (your frontend, API, etc.)
3. This automatically creates both:
   - App Registration (development/config)
   - Enterprise Application (deployment/access)

Scenario 2: Third-Party SaaS Apps

Added from Enterprise Apps:

1. Enterprise applications → New application
2. Browse Azure AD Gallery (e.g., Slack, Salesforce, Zoom)
3. Click Create to add to your tenant
4. Only creates an Enterprise Application (no App Registration needed)

Scenario 3: Custom Line-of-Business Apps

Uploaded/registered:

1. Developer creates app externally
2. Registers it via App registrations
3. Becomes available as Enterprise Application

---

How Groups Get Assigned to Applications

Method 1: From the Enterprise Application

1. Enterprise applications → Select app → Users and groups
2. Click Add user/group
3. Select the group
4. Assign a role (if app has app roles defined)
5. Click Assign

Method 2: From the Group (What You're Seeing)

1. Groups → Select group → Applications
2. Shows all apps this group is assigned to
3. This is a read-only view of assignments

Method 3: Programmatically

Via PowerShell, Graph API, or automated provisioning

---

Where to Manage These Application Assignments

To View All Details About an Application:

https://entra.microsoft.com → Enterprise applications → Find the app:

Key sections:

- Overview - Basic info, app type, visibility
- Properties - App settings, visibility, assignment required
- Users and groups - Who/which groups are assigned (⭐ This is where you manage assignments)
- Single sign-on - SSO configuration
- Permissions - API permissions granted
- Provisioning - Automatic user provisioning settings
- Application proxy - If using for on-prem apps

---

To Find All Apps a Specific Group Has Access To

From the group (what you're already doing):

1. Groups → Select group → Applications
2. Lists all directly assigned applications

This view shows:

- Application name
- Assignment type (Direct)
- Role assigned (if applicable)

---

Common Application Types You'll See

1. Microsoft Applications (built-in):
   - Office 365
   - Microsoft Teams
   - Azure Portal
   - Dynamics 365

2. Third-Party SaaS Apps:
   - Slack
   - Salesforce
   - Zoom
   - ServiceNow

3. Custom/Internal Apps:
   - Your company's web applications
   - Custom APIs
   - Line-of-business apps

4. Service Principals:
   - Managed identities
   - Azure services
   - Automation accounts

---

To Remove a Group Assignment from an Application

Option A: From Enterprise Application

1. Enterprise applications → Select app → Users and groups
2. Find the group
3. Check the box next to it
4. Click Remove (top toolbar)

Option B: Cannot Remove from Group View

- The Applications section in a group is read-only
- You must go to the Enterprise Application itself to remove

---

Quick Summary

To answer your questions:

Where created?

- Enterprise applications section (for third-party/SaaS apps)
- App registrations → then Enterprise App created automatically (for custom apps)

Where viewed?

- Enterprise applications → All applications → Search for app name
- Or click the app name from the group's Applications section (if it's a link)

Where managed?

- Enterprise applications → Select app → Users and groups (to manage assignments)

Primary location to work with these:

- https://entra.microsoft.com → Enterprise applications → All applications

> so I have an enterprise app but where can i see what roles are available in that app and how can I assign a user or groups that role

⏺ Where to See Available Roles in an Enterprise App

Option 1: Enterprise Application - Users and Groups Section

https://entra.microsoft.com:

1. Enterprise applications → All applications
2. Search for and click your application
3. Click Users and groups in the left menu
4. Click + Add user/group
5. Click Select a role (or it may say "None Selected")
6. This shows all available roles for the app

If it says "Default Access" or "No roles defined", the app doesn't have custom app roles.

---

Option 2: Check the App Registration (Where Roles Are Defined)

https://entra.microsoft.com:

1. App registrations → All applications
2. Search for the same application name
3. Click App roles in the left menu
4. Here you'll see all defined roles with:
   - Display name
   - Value (what appears in the token)
   - Description
   - Allowed member types (Users/Groups, Applications)

Note: Some Enterprise Apps (especially third-party SaaS apps) don't have a corresponding App Registration in your tenant - they're owned by the vendor.

---

Option 3: Enterprise Application Properties

In the Enterprise Application:

1. Click Properties in the left menu
2. Look for Application ID
3. Copy it and search for it in App registrations using the Application ID
4. Then check App roles section

---

How to Assign Users or Groups to Roles

Step-by-Step Assignment Process

https://entra.microsoft.com → Enterprise applications → Your app:

1. Click Users and groups in the left menu
2. Click + Add user/group (top toolbar)
3. Under Users and groups:
   - Click None Selected
   - Search for and select users or groups
   - Click Select

4. Under Select a role:
   - Click None Selected (or the role name shown)
   - You'll see all available roles for this app
   - Select the role you want to assign
   - Click Select

5. Click Assign button at the bottom

Done! The user/group now has that role in the application.

---

What If No Roles Are Available?

If you click "Select a role" and see:

- "No roles available" or
- Only "Default Access" or
- "msiam_access" (Microsoft default)

This means: The app doesn't have custom app roles defined.

Solutions:

If It's Your Own App (You Control It):

Create app roles:

1. App registrations → Your app → App roles
2. Click + Create app role
3. Fill in:
   - Display name: Admin, User, Editor, etc.
   - Allowed member types: Users/Groups, Applications, or Both
   - Value: admin, user, editor (appears in JWT token)
   - Description: What this role can do

4. Click Apply
5. The role is now available for assignment in Enterprise Application

If It's a Third-Party App:

- Check the vendor's documentation for available roles
- Some apps sync roles from the app itself
- Some don't support role-based assignments

---

Example: Complete Workflow

Scenario: Assign "IT Support" group the "Admin" role in your app

1. https://entra.microsoft.com → Enterprise applications → Find your app
2. Users and groups → + Add user/group
3. Users and groups section:
   - Click None Selected
   - Search for "IT Support"
   - Check the box next to "IT Support" group
   - Click Select

4. Select a role section:
   - Click None Selected
   - You'll see available roles like:
     - Admin
     - User
     - Editor
     - (etc.)
   - Click Admin
   - Click Select

5. Click Assign
6. Result: All members of "IT Support" group now have the "Admin" role for this application

---

Viewing Current Role Assignments

To see who has what role:

Enterprise applications → Your app → Users and groups

You'll see a table with:

- Name (user or group name)
- Type (User or Group)
- Assignment type (Direct or Inherited)
- Role (Admin, User, etc.)

---

Checking What Roles a Specific User/Group Has

From User Perspective:

1. Users → All users → Select user
2. Click Applications in left menu
3. Shows all apps and assigned roles

From Group Perspective:

1. Groups → Select group → Applications
2. Shows all apps and assigned roles for that group

---

Quick Reference

To see available roles:

- Enterprise applications → Your app → Users and groups → Add user/group → Select a role
- Or: App registrations → Your app → App roles

To assign a role:

1. Enterprise applications → Your app → Users and groups
2. - Add user/group
3. Select user/group
4. Select role
5. Click Assign

To view assignments:

- Enterprise applications → Your app → Users and groups (see the table)
