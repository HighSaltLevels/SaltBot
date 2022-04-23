""" Module for youtube related operations """
import os
from api import API, APIError

YOUTUBE_AUTH = os.getenv("YOUTUBE_AUTH")
YOUTUBE_MAX_IDX = 49
MAX_REQUESTED_VIDS = 50
BASE_URL = "https://www.youtube.com/watch?v="


class Youtube(API):
    """Represents a Youtube Object"""

    def get_video(self, idx):
        """Get the video at {idx}"""
        self.validate_num_vids()
        max_idx = len(self.data["items"])
        self.validate_idx(idx, max_idx)

        return BASE_URL + self.data["items"][idx]["id"]["videoId"]

    def validate_num_vids(self):
        """Ensure at least one video returned from youtube"""
        if self.num_videos == 0:
            raise APIError("```Sorry, there were no videos for that query :(```")

    @staticmethod
    def _create_url(query_args):
        """Create the youtube query string"""
        query = ",".join(query_args)
        return (
            f"https://www.googleapis.com/youtube/v3/search?key={YOUTUBE_AUTH}"
            f"&q={query}&maxResults={MAX_REQUESTED_VIDS}&type=video"
        )

    @property
    def num_videos(self):
        """Return the total number of returned videos"""
        return len(self.data["items"]) if self.response else 0
