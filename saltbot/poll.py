""" modlue for poll operations """
from http import HTTPStatus
import json
import uuid

from kubernetes.client.rest import ApiException

from common.k8s.configmap import ConfigMap

POLL_HELP_MSG = (
    "```How to set a poll:\nType the !poll command followed by the question, the "
    "answers, and the time all separated by semicolons. For Example:\n\n "
    "!poll How many times do you poop daily? ; Less than once ; Once ; Twice ; "
    "More than twice ; ends in 4 hours\n\nThe final poll expiry has to be in "
    'the format "ends in X Y" where "X" is any positive integer and "Y" '
    "is one of (hours, hour, minutes, minute, seconds, second)```"
)


class PollError(Exception):
    """
    Raised when there is an unknown poll error. It is expected
    to send the error string back to the user.
    """


# pylint: disable=too-many-instance-attributes
class Poll:
    """Poll class"""

    def __init__(self, **kwargs):
        # Set the expiry if it's passed in or use the time_length
        self._expiry = kwargs.get("expiry")
        self._time_length = kwargs.get("time_length")
        if self._time_length:
            self._expiry = self._time_length.timeout

        self.unique_id = kwargs.get("unique_id")
        self.author = kwargs.get("author")
        self.choices = kwargs.get("choices", [])
        self.votes = kwargs.get("votes")
        self.prompt = kwargs.get("prompt")
        self.channel = kwargs.get("channel")

    @property
    def data(self):
        """Return a json representation of the poll"""
        return {
            "unique_id": self.unique_id,
            "author": self.author,
            "kind": "poll",
            "claimed": "false",
            "prompt": self.prompt,
            # To be stored in a config map, it has to be a string
            "choices": json.dumps(self.choices),
            "expiry": self._expiry,
            "channel": self.channel,
            "votes": json.dumps(self.votes),
        }

    @property
    def expiry(self):
        """Read only poll expiry"""
        return self._expiry

    def create(self):
        """Create the configmap representation"""
        # Build an ID using the last part of a UUID
        self.unique_id = str(uuid.uuid4()).split("-")[-1]
        name = f"poll-{self.unique_id}"

        config_map = ConfigMap()
        config_map.create(name, labels={}, data=self.data)

        return self.unique_id

    def patch(self, poll_id):
        """Patch an existing poll object by ID"""
        name = f"poll-{poll_id}"
        config_map = ConfigMap()
        config_map.patch(name, self.data)

    @staticmethod
    def get(poll_id):
        """Get the corresponding poll by ID"""
        name = f"poll-{poll_id}"
        config_map = ConfigMap()
        try:
            return config_map.get(name)
        except ApiException as error:
            if error.status == HTTPStatus.NOT_FOUND:
                raise PollError("```Poll {poll_id} does not exist!```") from error
            raise error
