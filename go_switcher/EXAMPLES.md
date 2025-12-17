# Examples

## Basic Usage Examples

### 1. First Time Setup

When you run the tool for the first time, you need to provide your SSO start URL:

```bash
./bin/aws-switcher --start-url https://mycompany.awsapps.com/start/ --update
```

This will:
- Open your browser for AWS SSO authentication
- Fetch all available accounts and roles
- Save the configuration to `~/.aws/switcher-config.json`

### 2. Daily Usage

Once configured, simply run:

```bash
./bin/aws-switcher
```

This will show an interactive list like:
```
Production Account (123456789012) AdminRole
Development Account (123456789013) DevRole
Staging Account (123456789014) ReadOnlyRole
```

Use arrow keys or type to search, press Enter to select.

### 3. Updating Account List

When new accounts are added to your SSO or roles change:

```bash
./bin/aws-switcher --update
```

### 4. Using Different SSO Regions

```bash
./bin/aws-switcher --sso-region eu-west-1 --update
```

## Configuration Examples

### Minimal Configuration
```json
{
  "start_url": "https://mycompany.awsapps.com/start/",
  "sso_region": "us-east-1",
  "accounts": []
}
```

### Full Configuration
```json
{
  "start_url": "https://mycompany.awsapps.com/start/",
  "sso_region": "us-east-1",
  "accounts": [
    {
      "id": "123456789012",
      "name": "Production",
      "role_name": "AdminRole"
    },
    {
      "id": "123456789013", 
      "name": "Development",
      "role_name": "DevRole"
    },
    {
      "id": "123456789014",
      "name": "Staging", 
      "role_name": "ReadOnlyRole"
    }
  ]
}
```

## AWS Credentials File

After selecting an account, your `~/.aws/credentials` will look like:

```ini
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
aws_session_token = AQoDYXdzEJr...example...token
```

## Integration with AWS CLI

After running the switcher, you can immediately use AWS CLI:

```bash
# Switch account
./bin/aws-switcher

# Use AWS CLI with the selected account
aws s3 ls
aws ec2 describe-instances
aws iam get-user
```

## Troubleshooting

### Browser doesn't open automatically
If the browser doesn't open automatically during SSO login, copy the URL from the terminal output and open it manually.

### Credentials not working
- Check that the session hasn't expired
- Try running the switcher again to refresh credentials
- Verify your SSO access hasn't been revoked

### No accounts found
- Ensure your SSO start URL is correct
- Verify you have access to AWS accounts through SSO
- Try updating with: `./bin/aws-switcher --update`
