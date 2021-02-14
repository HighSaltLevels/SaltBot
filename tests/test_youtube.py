""" Youtube Test Module """

from unittest import mock

from requests.models import Response
import pytest

from api import APIError
from youtube import Youtube, YOUTUBE_AUTH, MAX_REQUESTED_VIDS, BASE_URL


@pytest.fixture(name="mock_response")
def create_mock_response():
    """ Create a mock requests response object """
    resp = mock.Mock(spec=Response)
    resp.status_code = 200
    resp.json.return_value = {"items": [{"id": {"videoId": "isaac"}}]}
    return resp


@mock.patch("api.API._request")
def test_get_video(mock_request, mock_response):
    """ Test getting a YT video """
    mock_request.return_value = mock_response
    youtube = Youtube("northernlion")
    assert youtube.url == (
        f"https://www.googleapis.com/youtube/v3/search"
        f"?key={YOUTUBE_AUTH}&q=northernlion&maxResults"
        f"={MAX_REQUESTED_VIDS}&type=video"
    )

    assert youtube.get_video(0) == f"{BASE_URL}isaac"

    # Test getting a youtube video but no results
    mock_response.json.return_value = {"items": []}
    youtube = Youtube("northernlion")
    with pytest.raises(APIError) as error:
        youtube.validate_num_vids()

    assert "Sorry, there were no videos" in str(error)
