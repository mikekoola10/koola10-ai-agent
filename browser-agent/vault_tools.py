import requests
from urllib.parse import quote

VAULT_URL = "https://script.google.com/macros/s/AKfycbxFFviECwZcWEZy9HLo2aEUlAqB-brL5MZFcn1OtTe8wYurw4G7AJltd5dHAE6bQRRg/exec"

def add_vault_entry(description: str, amount: float, type: str, notes: str = ""):
    """
    Sends a transaction entry to the Google Sheets vault.
    :param description: Description of the transaction.
    :param amount: Amount of the transaction.
    :param type: Type of transaction (income, expense, trade, profit).
    :param notes: Optional notes.
    :return: Boolean indicating success.
    """
    params = f"?action=add&description={quote(description)}&amount={amount:.2f}&type={quote(type)}&notes={quote(notes)}"
    try:
        resp = requests.get(VAULT_URL + params, timeout=10)
        return resp.ok
    except Exception as e:
        print(f"Error adding vault entry: {e}")
        return False

def get_vault_summary():
    """
    Retrieves the vault summary.
    :return: Summary string or None if failed.
    """
    params = "?action=summary"
    try:
        resp = requests.get(VAULT_URL + params, timeout=10)
        if resp.ok:
            return resp.text
        return None
    except Exception as e:
        print(f"Error getting vault summary: {e}")
        return None
