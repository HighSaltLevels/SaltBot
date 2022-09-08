""" Messenger module for sending discord messages """

import json

from common.logger import LOGGER


class Messenger:
    """Messenger class"""

    def __init__(self, client):
        self._client = client

    async def send_poll_results(self, poll_spec):
        """Send the results of the poll"""
        total_votes = 0
        results = {}
        choices = json.loads(poll_spec["choices"])
        votes = json.loads(poll_spec["votes"])

        for choice_num, _ in enumerate(choices):
            total_for_this_choice = len(votes[str(choice_num)])
            results[choice_num] = total_for_this_choice
            total_votes += total_for_this_choice

        msg = f"```Results (Total votes: {total_votes}):\n\n"
        try:
            for result in results:
                choice = choices[result]
                msg += (
                    f"\t{choice} -> "
                    f"{float(len(votes[str(result)])/total_votes) * 100:.0f}"
                    "%\n"
                )
        except ZeroDivisionError:
            msg = "```No one voted on this poll :("

        await self.send(f"{msg}```", poll_spec)

    async def send(self, msg, spec):
        """Send the message to the channel or author"""
        LOGGER.log(
            f"Attempting to send reminder {spec['unique_id']} to {spec['channel']}"
        )
        channel = self._client.get_channel(int(spec["channel"]))
        if channel is None:
            LOGGER.log("It's a DM, look up the channel")
            for guild in self._client.guilds:
                for member in guild.members:
                    if member.id == int(spec["author"]):
                        await member.send(msg)

        else:
            await channel.send(msg)
