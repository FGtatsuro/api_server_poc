import os

import requests

API_URL = os.environ.get('API_URL') or 'http://127.0.0.1:8080'


def test_smoke_access():
    path = '/api'
    resp = requests.get(
        f'{API_URL}{path}',
        headers={'authorization': 'api_token'}
    )
    assert resp.status_code == 200
