""" API Test Module """

from unittest import mock

import pytest

from api import APIError, API


def test_create_url():
    """ Test creating the url is not implemented for the base class """
    with pytest.raises(NotImplementedError):
        API("foo")


def test_request():
    """ Test making a request on the url """
    with mock.patch("api.API._create_url") as mock_create:
        mock_create.return_value = "https://google.com"
        with mock.patch("requests.get") as mock_get:
            with mock.patch("api.API.validate_status"):
                API("foo")
                mock_get.assert_called()


def test_validate_idx():
    """ Test the validation of the response index """
    with mock.patch("api.API._request"):
        with mock.patch("api.API._create_url"):
            with mock.patch("api.API.validate_status"):
                api = API("foo")

                # Test not an integer specified
                with pytest.raises(APIError) as error:
                    api.validate_idx("foo", "bar")
                assert "You have to specify an integer" in str(error)

                # Test index above valid range
                with pytest.raises(APIError) as error:
                    api.validate_idx(idx=5, max_idx=4)
                assert "The index must be between 0 and" in str(error)

                # Test index below valid range
                with pytest.raises(APIError) as error:
                    api.validate_idx(idx=-1, max_idx=4)
                assert "The index must be between 0 and" in str(error)
