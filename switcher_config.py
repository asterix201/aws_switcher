# A script to login to AWS accounts and update AWS creds whithout requiring to copy/paste from Control Tower SSO page

from email.policy import default
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

    Returns:
        dict: { 'account_id': account_id, 'role_name': role_name }
    """
    longest_name_len = max([len(acc[1]) for acc in accounts])
    account_choices = [
        f"{acc[1]} {' '*(longest_name_len - len(acc[1]))} {acc[0]} {acc[2]}"
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

    account_name = re.search(r"^.*?(?=\b\d{12}\b)", answers["user_option"]).group(), # get account name
    account_id = re.search(r"\b\d{12}\b", answers["user_option"]).group(), # 12 digits, to get aws account id
    role_name = re.search(r"\b\w+\b$", answers["user_option"]).group(), # last word, to get role name

    return {
        "account_name": account_name[0].strip(),
        "account_id": account_id[0].strip(),
        "role_name": role_name[0].strip(),
    }


def read_aws_config(aws_config: str = "~/.aws/config") -> dict:
    accounts = {}
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
            accounts[profile_name] = {
                "sso_start_url": config[section]["sso_start_url"],
                "sso_region": config[section]["sso_region"],
                "sso_account_id": config[section]["sso_account_id"],
                "sso_role_name": config[section]["sso_role_name"],
                "region": config[section]["region"],
                "profile_name": profile_name,
            }
    return accounts


def update_aws_config(
    aws_config: str, sso_start_url: str, sso_region: str, region: str
):
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
    if update:
        if not sso_start_url:
            raise click.ClickException("No --sso_start_url specified")
        print(aws_config, sso_start_url, sso_region, region)
        update_aws_config(aws_config, sso_start_url, sso_region, region)
    accounts = read_aws_config()
    accounts_formatted = [
        (account["sso_account_id"], account["profile_name"], account["sso_role_name"])
        for account in accounts.values()
    ]
    account_role = choose_account_role(accounts_formatted)
    account_name, account_id, role_name = account_role["account_name"], account_role["account_id"], account_role["role_name"]
    print(f"profile_name: {account_name}")
    print(f"account_id:   {account_id}")
    print(f"role_name:    {role_name}")
    session = login(
        start_url=accounts[account_name]["sso_start_url"],
        sso_region=accounts[account_name]["sso_region"],
    )
    role_credentials = get_role_credentials(
        accessToken=session["accessToken"], account_id=account_id, role_name=role_name
    )
    write_creds(role_credentials)


if __name__ == "__main__":
    main()
