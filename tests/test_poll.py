""" Poll Test Module """

import os

import pytest

from poll import Poll, POLL_DIR, POLL_HELP_MSG
from timelength import TimeLength


def test_poll_save_load(capsys):
    """ Test saving and loading a poll """

    # Test saving a poll
    poll = Poll(
        time_length=TimeLength("hours", 5), votes=[], prompt="foo", channel_id=1234
    )

    poll.save()
    assert os.path.isfile(f"{POLL_DIR}/{poll.poll_id}.json")

    # Test Loading a poll
    other_poll = Poll()
    other_poll.load(poll.poll_id)

    assert poll.poll_id == other_poll.poll_id
    assert poll.expiry == other_poll.expiry
    assert poll.prompt == other_poll.prompt
    assert poll.choices == other_poll.choices
    assert poll.votes == other_poll.votes
    assert poll.channel_id == other_poll.channel_id

    # Test error loading poll
    with open(f"{POLL_DIR}/1234.json", "w") as stream:
        stream.write("{}")
    with pytest.raises(ValueError) as error:
        other_poll.load("1234")

    assert f"Something went wrong with poll 1234" in str(error)

    # Test delete
    poll.delete()
    other_poll.delete()
    assert not os.path.isfile(poll.poll_id)

    poll.delete()
    assert f"Warning: {poll.poll_id} does not exist" in capsys.readouterr().out
