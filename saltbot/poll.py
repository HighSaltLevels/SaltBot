""" modlue for poll operations """
import asyncio
import glob
import json
import os
from pathlib import Path
import time

POLL_DIR = os.path.join(str(Path.home()), ".config/saltbot/polls")
os.makedirs(POLL_DIR, exist_ok=True)

POLL_HELP_MSG = (
    "```How to set a poll:\nType the !poll command followed by the question, the "
    "answers, and the time all separated by semicolons. For Example:\n\n "
    "!poll How many times do you poop daily? ; Less than once ; Once ; Twice ; "
    "More than twice ; ends in 4 hours\n\nThe final poll expiry has to be in "
    'the format "ends in X Y" where "X" is any positive integer and "Y" '
    "is one of (hours, hour, minutes, minute, seconds, second)```"
)


class Poll:
    """ Poll class """

    def __init__(self, **kwargs):
        self._expiry = None
        self._poll_id = None
        self._time_length = kwargs.get("time_length")
        if self._time_length:
            self._expiry = self._time_length.timeout
            self._poll_id = self._time_length.unique_id

        # Default to no choices
        self.choices = kwargs.get("choices", 0)
        self.votes = kwargs.get("votes")

        self.prompt = kwargs.get("prompt")
        self.channel_id = kwargs.get("channel_id")

    @property
    def poll_id(self):
        """ poll_id read only property """
        return self._poll_id

    @property
    def expiry(self):
        """ Read only poll expiry """
        return self._expiry

    def save(self):
        """ Save the poll to disk """
        data = {
            "prompt": self.prompt,
            "choices": self.choices,
            "expiry": self._expiry,
            "poll_id": self._poll_id,
            "channel_id": self.channel_id,
            "votes": self.votes,
        }
        with open(f"{POLL_DIR}/{self._poll_id}.json", "w") as stream:
            stream.write(json.dumps(data))

    def load(self, poll_id):
        """ Load the data from disk """
        with open(f"{POLL_DIR}/{poll_id}.json") as stream:
            data = json.load(stream)

        try:
            self.prompt = data["prompt"]
            self.choices = data["choices"]
            self._expiry = data["expiry"]
            self._poll_id = data["poll_id"]
            self.channel_id = data["channel_id"]
            self.votes = data["votes"]
        except KeyError as error:
            raise ValueError(
                f"```Something went wrong with poll {poll_id}```"
            ) from error

    def delete(self):
        """ Delete the poll """
        # Poll ID has to exist to be able to delete a poll
        assert self._poll_id is not None
        try:
            os.remove(f"{POLL_DIR}/{self._poll_id}.json")
        except FileNotFoundError:
            print(f"Warning: {self._poll_id} does not exist!", flush=True)


async def monitor_polls(discord_client):
    """ Check poll files for expiry every 5 seconds """
    while True:
        poll = Poll()
        poll_files = glob.glob(f"{POLL_DIR}/*")
        for poll_file in poll_files:
            # Get the ID from the file name
            poll_id = os.path.splitext(poll_file)[0].split("/")[-1]
            poll.load(poll_id)

            if time.time() > poll.expiry:
                channel = discord_client.get_channel(poll.channel_id)
                total_votes = 0
                results = {}
                for choice_num in range(len(poll.choices)):
                    total_for_this_choice = len(poll.votes[str(choice_num)])
                    results[choice_num] = total_for_this_choice
                    total_votes += total_for_this_choice

                response = f"```Results (Total votes: {total_votes}):\n\n"
                try:
                    for result in results:
                        choice = poll.choices[result]
                        response += (
                            f"\t{choice} -> "
                            f"{float(len(poll.votes[str(result)])/total_votes) * 100:.0f}"
                            "%\n"
                        )
                except ZeroDivisionError:
                    response = "```No one voted on this poll :("

                await channel.send(f"{response}```")
                poll.delete()

        await asyncio.sleep(5)
