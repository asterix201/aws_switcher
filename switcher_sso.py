# A script to login to AWS accounts and update AWS creds whithout requiring to copy/paste from Control Tower SSO page

import aws_sso_lib
import boto3
import configparser
from InquirerPy import prompt
import os

def login(start_url: str, sso_region: str) -> dict:
    """Login to AWS with SSO account

    Args:
        start_url (str): SSO start URL.
        sso_region (str): SSO region.

    Returns:
        dict: Returns the token dict as returned by sso-oidc:CreateToken, which contains the actual authorization token, as well as the expiration
    """
    return aws_sso_lib.login(
        start_url=start_url,
        sso_region=sso_region,
    )


def get_role_credentials(accessToken: dict, account_id: str, role_name: str,
    region: str="us-east-1") -> dict:
    """Create sso client for the given account and return temporary aws credentials

    Args:
        accessToken (dict): The token issued by the CreateToken API call
        account_id (str): aws account number
        role_name (str): aws role account name
        region (str, optional): Region for API calls. Defaults to "us-east-1".

    Returns:
        dict: Dict with temporary aws credentials
    """
    client = boto3.client('sso',region_name=region)
    response = client.get_role_credentials(
        roleName=role_name,
        accountId=account_id,
        accessToken=accessToken
    )

    return response['roleCredentials']


def write_creds(roleCredentials: dict, aws_credentials: str='~/.aws/credentials', aws_config: str='~/.aws/config'):
    """Save aws credentials to corresponding config files. 
       Add default profile if it not already present in the config file.

    Args:
        roleCredentials (dict): roleCredentials
        aws_credentials (str, optional): Configuration file. Defaults to '~/.aws/credentials'.
        aws_config (str, optional): Credential file. Defaults to '~/.aws/config'.
    """
    credentials = configparser.ConfigParser()
    credentials['default'] = {
        "aws_access_key_id": roleCredentials['accessKeyId'],
        "aws_secret_access_key": roleCredentials['secretAccessKey'],
        "aws_session_token": roleCredentials['sessionToken'],
    }
    with open(os.path.expanduser(aws_credentials), 'w') as credentials_file:
        credentials.write(credentials_file)

    config = configparser.ConfigParser()
    config.read_file(open(os.path.expanduser(aws_config)))
    if not config.has_section('profile default'):
        config.add_section('profile default')
        with open(os.path.expanduser(aws_config), 'w') as config_file:
            config.write(config_file)


def choose_account_role(accounts: list) -> dict:
    """Ask to choose an account with interactive cli

    Args:
        accounts (list): List of account in tuples in format [(account_id, account_name, role_name)].

    Returns:
        dict: { 'account_id': account_id, 'role_name': role_name }
    """
    longets_name_len = max([len(acc[1]) for acc in accounts])
    account_choices = [ f"{acc[1]} {' '*(longets_name_len - len(acc[1]))} {acc[0]} {acc[2]}" for acc in accounts ]
    questions = [
        {
            'type': 'fuzzy',
            'name': 'user_option',
            'message': 'choose an account',
            "max_height": "70%",
            'choices': account_choices,
        }
    ]
    answers = prompt(questions)
    
    return {'account_id': answers["user_option"].split()[1], 
            'role_name': answers["user_option"].split()[2]}


def main():
    start_url = "https://my-sso-portal.awsapps.com/start/"
    sso_region = "us-east-1"
    accounts = [ account for account in aws_sso_lib.list_available_roles(start_url=start_url, sso_region=sso_region, login=True)] 
    account_role = choose_account_role(accounts)
    account_id, role_name = account_role["account_id"], account_role["role_name"]
    print("account_id: ", account_id)
    print("role_name: ", role_name)
    session = login(start_url=start_url, sso_region=sso_region)
    role_credentials = get_role_credentials(accessToken=session['accessToken'], account_id=account_id, role_name=role_name)
    write_creds(role_credentials)

if __name__ == "__main__":
    main()



