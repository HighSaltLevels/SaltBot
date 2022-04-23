""" Giphy interface module """
import os

from api import API, APIError

GIPHY_AUTH = os.getenv("GIPHY_AUTH")


class Giphy(API):
    """Object for interfacing with Giphy"""

    @property
    def num_gifs(self):
        """Get the total number of gifs"""
        return len(self.data["data"]) if self.response else 0

    @property
    def all_gifs(self):
        """Return a list of all gifs"""
        self.validate_num_gifs()
        return [
            self.data["data"][index]["bitly_gif_url"] for index in range(self.num_gifs)
        ]

    def get_gif(self, idx):
        """Get the gif at {idx}"""
        self.validate_num_gifs()
        max_idx = len(self.data["data"])
        self.validate_idx(idx, max_idx)
        return self.data["data"][idx]["bitly_gif_url"]

    @staticmethod
    def _create_url(query_args):
        """Create the url query url"""
        query = "+".join(query_args)
        return f"http://api.giphy.com/v1/gifs/search?q={query}&api_key={GIPHY_AUTH}"

    def validate_num_gifs(self):
        """Ensure there was enough gifs for that query"""
        if self.num_gifs == 0:
            raise APIError("```Sorry, there were no gifs for that query :(```")
