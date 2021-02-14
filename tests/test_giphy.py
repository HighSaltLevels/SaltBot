""" Giphy Test Module """

from unittest import mock

import pytest
from requests.models import Response

from api import APIError
from giphy import Giphy, GIPHY_AUTH


@pytest.fixture(name="mock_response")
def create_mock_response():
    """ Create a mock requests response object """
    resp = mock.Mock(spec=Response)
    resp.status_code = 200
    resp.json.return_value = {"data": [{"bitly_gif_url": "foo"}]}
    return resp


@mock.patch("api.API._request")
def test_request_gif(mock_request, mock_response):
    """ Test getting a gif """
    # Try getting a good gif
    mock_request.return_value = mock_response
    giphy = Giphy("dog")
    assert (
        giphy.url == f"http://api.giphy.com/v1/gifs/search?q=dog&api_key={GIPHY_AUTH}"
    )

    # Try getting a gif but bad status code
    for bad_code in (400, 404, 500):
        mock_response.status_code = bad_code
        with pytest.raises(APIError) as error:
            giphy = Giphy("dog")

        assert "Sorry, I had trouble getting that query" in str(error)


@mock.patch("api.API._request")
def test_giphy_properties(mock_request, mock_response):
    """ Test the giphy {num_gifs} and {all_gifs} properties """
    mock_request.return_value = mock_response

    # Mock response has 1 "gif"
    giphy = Giphy("dog")
    assert giphy.num_gifs == 1
    assert giphy.all_gifs == ["foo"]


@mock.patch("api.API._request")
def test_get_gif(mock_request, mock_response):
    """ Test getting a gif and retrieving from the json resp """
    mock_request.return_value = mock_response

    giphy = Giphy("dog")
    assert giphy.get_gif(0) == "foo"


@mock.patch("api.API._request")
def test_validate_num_gifs(mock_request, mock_response):
    """ Test validation of the number of gifs in a resp """
    # Test a non-zero number of gifs
    mock_request.return_value = mock_response
    giphy = Giphy("dog")
    giphy.validate_num_gifs()

    # Test a zero number of gifs
    mock_response.json.return_value = {"data": []}
    giphy = Giphy("dog")
    with pytest.raises(APIError) as error:
        giphy.validate_num_gifs()

    assert "Sorry, there were no gifs" in str(error)
