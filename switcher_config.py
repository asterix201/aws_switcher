# A script to login to AWS accounts and update AWS creds whithout requiring to copy/paste from Control Tower SSO page

import aws_sso_lib
import boto3
import configparser
from InquirerPy import prompt
import os
from typing import Tuple
import click
import re


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


def get_role_credentials(
    accessToken: dict, account_id: str, role_name: str, region: str = "us-east-1"
) -> dict:
    """Create sso client for the given account and return temporary aws credentials

    Args:
        accessToken (dict): The token issued by the CreateToken API call
        account_id (str): aws account number
        role_name (str): aws role account name
        region (str, optional): Region for API calls. Defaults to "us-east-1".

    Returns:
        dict: Dict with temporary aws credentials
    """
    client = boto3.client("sso", region_name=region)
    response = client.get_role_credentials(
        roleName=role_name, accountId=account_id, accessToken=accessToken
    )

    return response["roleCredentials"]


def write_creds(
    roleCredentials: dict,
    aws_credentials: str = "~/.aws/credentials",
    aws_config: str = "~/.aws/config",
):
    """Save aws credentials to corresponding config files.
       Add default profile if it not already present in the config file.

    Args:
        roleCredentials (dict): roleCredentials
        aws_credentials (str, optional): Configuration file. Defaults to '~/.aws/credentials'.
        aws_config (str, optional): Credential file. Defaults to '~/.aws/config'.
    """
    credentials = configparser.ConfigParser()
    credentials["default"] = {
        "aws_access_key_id": roleCredentials["accessKeyId"],
        "aws_secret_access_key": roleCredentials["secretAccessKey"],
        "aws_session_token": roleCredentials["sessionToken"],
    }
    with open(os.path.expanduser(aws_credentials), "w") as credentials_file:
        credentials.write(credentials_file)

    config = configparser.ConfigParser()
    config.read_file(open(os.path.expanduser(aws_config)))
    if not config.has_section("profile default"):
        config.add_section("profile default")
        with open(os.path.expanduser(aws_config), "w") as config_file:
            config.write(config_file)


def choose_account_role(accounts: list) -> dict:
    """Ask to choose an account with interactive cli

    Args:
        accounts (list): List of account in tuples in format [(account_id, account_name, role_name)].
        [(sso_account_id,  profile_name, sso_role_name)]

    Returns:
        dict: { 'account_id': account_id, 'role_name': role_name }
    """
    longest_name_len = max([len(acc[1]) for acc in accounts])
    longest_role_len = max([len(acc[2]) for acc in accounts])
    account_choices = [
        f"{acc[1]} {' '*(longest_name_len - len(acc[1]))} {acc[2]} {' '*(longest_role_len - len(acc[2]))} {acc[0]}"
        for acc in accounts
    ]
    questions = [
        {
            "type": "fuzzy",
            "name": "user_option",
            "message": "choose an account",
            "max_height": "70%",
            "choices": account_choices,
        }
    ]
    answers = prompt(questions)

    account_name = answers["user_option"].split()[:-2] # get list of words, except last two - profile account name
    account_id = answers["user_option"].split()[-1] # last word, to get aws account id
    role_name = answers["user_option"].split()[-2] # get role name

    return {
        "account_name": " ".join(account_name).strip(),
        "account_id": account_id.strip(),
        "role_name": role_name.strip(),
    }


def read_aws_config(aws_config: str = "~/.aws/config") -> list:
    """Read aws config file and return a list of accounts

    Args:
        aws_config (str, optional): path to aws config file. Defaults to "~/.aws/config".

    Returns:
        list: List of accounts in format
    """
    accounts = []
    config = configparser.ConfigParser()
    config.read_file(open(os.path.expanduser(aws_config)))
    for section in config.sections():
        if all(
            config.has_option(section, option)
            for option in [
                "sso_start_url",
                "sso_region",
                "sso_account_id",
                "sso_role_name",
                "region",
            ]
        ):
            profile_name = section.split(' ', 1)[1]
            accounts.append({
                "sso_start_url": config[section]["sso_start_url"],
                "sso_region": config[section]["sso_region"],
                "sso_account_id": config[section]["sso_account_id"],
                "sso_role_name": config[section]["sso_role_name"],
                "region": config[section]["region"],
                "profile_name": profile_name,
            })
    return accounts


def update_aws_config(
    aws_config: str, sso_start_url: str, sso_region: str, region: str
):
    """Update aws config file with accessible accounts/roles

    Args:
        aws_config (str): path to aws config file. 
        sso_start_url (str): URL that points to the organization's AWS SSO user portal.
        sso_region (str): The AWS Region that contains the AWS SSO portal host.
        region (str): AWS Region to send requests to for commands requested using this profile.
    """
    accounts = [
        account
        for account in aws_sso_lib.list_available_roles(
            start_url=sso_start_url, sso_region=sso_region, login=True
        )
    ]
    config = configparser.ConfigParser()
    config.read_file(open(os.path.expanduser(aws_config)))
    for account, account_name, role_name in accounts:
        section = f"profile {account_name} {role_name[:7]}"
        if not config.has_section(section):
            config.add_section(section)
            config.set(section, "sso_start_url", sso_start_url)
            config.set(section, "sso_region", sso_region)
            config.set(section, "sso_account_id", account)
            config.set(section, "sso_role_name", role_name)
            config.set(section, "region", region)
            config.set(section, "output", "text")
    with open(os.path.expanduser(aws_config), "w") as config_file:
        config.write(config_file)


def check_aws_config(aws_config: str = "~/.aws/config"):
    """Check if aws config file is exist and has required permissions. Create aws_config file if not exist.

    Args:
        aws_config (str, optional): path to aws config file. Defaults to "~/.aws/config".
    """
    if not os.path.exists(os.path.expanduser(aws_config)):
        # check and create folders path to aws config file
        if not os.path.exists(os.path.dirname(os.path.expanduser(aws_config))):
            os.makedirs(os.path.dirname(os.path.expanduser(aws_config)))
            with open(os.path.expanduser(aws_config), "w") as config_file:
                config_file.write("[default]\n")
                config_file.write("region = us-east-1\n")
                config_file.write("output = text\n")


@click.command()
@click.option(
    "--update",
    is_flag=True,
    default=False,
    help="Update aws config file with accessible accounts/roles",
)
@click.option(
    "--sso_start_url",
    help="URL that points to the organization's AWS SSO user portal. Required if use --update flag",
)
@click.option(
    "--aws_config",
    default="~/.aws/config",
    help="AWS configuration file location. Defaults to '~/.aws/config",
)
@click.option(
    "--sso_region",
    default="us-east-1",
    help="The AWS Region that contains the AWS SSO portal host. Default 'us-east-1'",
)
@click.option(
    "--region",
    default="us-east-1",
    help="AWS Region to send requests to for commands requested using this profile. Default 'us-east-1'",
)


def main(update, aws_config, sso_start_url, sso_region, region):
    if not os.path.exists(os.path.expanduser(aws_config)): update=True
    if update:
        if not sso_start_url:
            raise click.ClickException("Updating aws config. No --sso_start_url specified")
        check_aws_config(aws_config)
        print(aws_config, sso_start_url, sso_region, region)
        update_aws_config(aws_config, sso_start_url, sso_region, region)
    accounts = read_aws_config()
    accounts_formatted = [
        (account["sso_account_id"], account["profile_name"], account["sso_role_name"])
        for account in accounts
    ]
    account_role = choose_account_role(accounts_formatted)
    account_name, account_id, role_name = account_role["account_name"], account_role["account_id"], account_role["role_name"]
    print(f"profile_name: {account_name}")
    print(f"account_id:   {account_id}")
    print(f"role_name:    {role_name}")

    # get profile config from accounts list of dicts with profiles
    login_account = [account for account in accounts if account["profile_name"] == account_name]
    session = login(
        start_url=login_account[0]["sso_start_url"],
        sso_region=login_account[0]["sso_region"],
    )
    role_credentials = get_role_credentials(
        accessToken=session["accessToken"], account_id=account_id, role_name=role_name
    )
    write_creds(role_credentials)


if __name__ == "__main__":
    main()
    
