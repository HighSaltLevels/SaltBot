"""
    API Base Module
"""
from http import HTTPStatus
import requests


class APIError(Exception):
    """ Raised if there is an issue getting the API query """


class API:
    """ API Base Class """

    def __init__(self, *query_args):
        self.url = self._create_url(list(query_args))
        self.response = self._request()
        self.data = self.response.json()
        self.validate_status()

    def _create_url(self, query_args):
        raise NotImplementedError

    def _request(self):
        """ Do an HTTP GET on the url """
        return requests.get(self.url)

    def validate_status(self):
        """ Check if status code was 200 """
        if self.response.status_code != HTTPStatus.OK:
            raise APIError("```Sorry, I had trouble getting that query :(```")

    @staticmethod
    def validate_idx(idx, max_idx):
        """ Verify that idx is valid and not out of range """
        # Specified index wasn't an integer
        try:
            idx = int(idx)
        except ValueError as error:
            raise APIError(
                "```You have to specify an integer if you want query by index!```"
            ) from error

        if idx < 0 or idx > max_idx:
            raise APIError(
                f"```The index must be between 0 and {max_idx} for this query```"
            )
