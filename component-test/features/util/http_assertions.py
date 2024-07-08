from hamcrest import assert_that, equal_to

from requests import Response


def assert_status_in(http_response: Response, status_from, status_to):
    method = http_response.request.method
    url = http_response.request.url
    status = http_response.status_code
    message = (
        f"Expected {method} {url} to have a status code of range "
        f"[{status_from}, {status_to}] but was {status}."
    )
    if not status_from <= http_response.status_code <= status_to:
        message += " The response was: \r\n" + str(http_response.content)
    assert status_from <= http_response.status_code <= status_to, message


def assert_status(http_response: Response, expected_status: int):
    method = http_response.request.method
    url = http_response.request.url
    status = http_response.status_code
    message = (
        f"Expected {method} {url}{http_response} to have a status code "
        f"{expected_status} but was {status}."
    )
    if not expected_status == http_response.status_code:
        message += " The response was: \r\n" + str(http_response.content)
    assert status == expected_status, message


def assert_2xx(http_response: Response):
    assert_status_in(http_response, 200, 299)


def assert_3xx(http_response: Response):
    assert_status_in(http_response, 300, 399)


def assert_4xx(http_response: Response):
    assert_status_in(http_response, 400, 499)


def assert_5xx(http_response: Response):
    assert_status_in(http_response, 500, 599)


def assert_error(http_response: Response, status, error_message):
    assert_status(http_response, status)
    response_body = http_response.json()
    assert_that(
        response_body["error"]["messaging"],
        equal_to(error_message),
    )
