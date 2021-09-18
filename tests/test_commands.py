""" Command Test Module """

from unittest import mock

import discord
import pytest

from commands import Command, MSG_DICT
from poll import Poll, POLL_HELP_MSG
from timelength import TimeLength
from reminder import REMINDER_HELP_MSG
from .utils import create_user_msg, create_mock_response


@pytest.fixture(name="poll")
def create_poll():
    """ Create a poll object for testing """
    time_length = TimeLength("years", 5)
    kwargs = {
        "time_length": time_length,
        "choices": ["a", "b", "c"],
        "votes": {"1": [], "2": [], "3": []},
        "prompt": "foo",
        "channel_id": "bar",
    }
    poll = Poll(**kwargs)
    poll.save()
    yield poll
    poll.delete()


def test_help():
    """ Test the help command """
    user_msg = create_user_msg()
    # Test all commands are documented in help
    for _help in {"!h", "!help"}:
        _, msg = Command(user_msg).commands[_help]()
        for key, value in MSG_DICT.items():
            assert key in msg
            assert value in msg


def test_remind():
    """ Test the remind command """
    user_msg = create_user_msg()
    # Test invalid reminder
    for reminder in {"!r", "!remind"}:
        # Test passing in no args or help command
        _, msg = Command(user_msg).commands[reminder]()
        assert msg == REMINDER_HELP_MSG
        _, msg = Command(user_msg).commands[reminder]("help")
        assert msg == REMINDER_HELP_MSG

        # Test reminder error is raised
        _, msg = Command(user_msg).commands[reminder]("set", "foo", "in", "five", "bar")
        assert "Invalid unit bar" in msg

        # Test create a normal reminder
        _, msg = Command(user_msg).commands[reminder](
            "set", "foo", "in", "5", "seconds"
        )
        assert ': "foo" at' in msg


def test_vote(poll):
    """ Test the vote command """
    user_msg = create_user_msg()
    for vote in {"!v", "!vote"}:
        # Test invalid vote command
        _, msg = Command(user_msg).commands[vote]()
        assert "Please format your vote as" in msg

        # Test poll not found
        _, msg = Command(user_msg).commands[vote]("fake-id", 3)
        assert "Poll fake-id does not exist or has expired" in msg

        # Test invalid poll prints out exception message instead of letting exception bubble up
        with mock.patch("poll.Poll.load") as m_load:
            m_load.side_effect = ValueError("mock value error")
            _, msg = Command(user_msg).commands[vote]("fake-id", 3)
            assert "mock value error" in msg

        # Test invalid vote choice
        _, msg = Command(user_msg).commands[vote](poll.poll_id, 420)
        assert "420 is not an available selection" in msg

        # Test valid vote choice
        _, msg = Command(user_msg).commands[vote](poll.poll_id, "2")
        assert "You have selected" in msg


def test_poll():
    """ Test the poll command """
    user_msg = create_user_msg()
    for cmd in {"!p", "!poll"}:
        # Test help queries
        _, msg = Command(user_msg).commands[cmd]()
        assert POLL_HELP_MSG in msg
        _, msg = Command(user_msg).commands[cmd]("help")
        assert POLL_HELP_MSG in msg

        # Test invalid polls
        user_msg = create_user_msg("foo ; bar ; ends in 4 blah")
        _, msg = Command(user_msg).commands[cmd]("blah")
        assert POLL_HELP_MSG in msg

        # Test valid poll
        user_msg = create_user_msg("foo ; bar ; ends in 4 hours")
        _, msg = Command(user_msg).commands[cmd]("blah")
        assert 'Type or DM me "!vote' in msg


def test_jeopardy():
    """ Test the jeopardy command """
    user_msg = create_user_msg()
    for cmd in {"!j", "!jeopardy"}:
        with mock.patch("commands.requests.get") as m_get:
            # Test server fails
            m_get.return_value = create_mock_response(status_code=500, kind="jeopardy")
            _, msg = Command(user_msg).commands[cmd]()
            assert "Something went wrong" in msg

        with mock.patch("commands.requests.get") as m_get:
            # Test server responds
            m_get.return_value = create_mock_response(status_code=200, kind="jeopardy")
            _, msg = Command(user_msg).commands[cmd]()
            assert "The Category is:" in msg


def test_whisper():
    """ Test the whisper command """
    user_msg = create_user_msg()
    for cmd in {"!pm", "!whisper"}:
        _, msg = Command(user_msg).commands[cmd]()
        assert "Hello foo" in msg


@mock.patch("api.API._request")
def test_gif(m_request):
    """ Test the gif command """
    user_msg = create_user_msg()
    for cmd in {"!g", "!gif"}:
        # Test invalid args
        _, msg = Command(user_msg).commands[cmd]()
        assert 'You have to type "!gif <query>"' in msg

        # Test with -a argument
        # This will fail because user_msg.channel is not a channel
        user_msg.channel = "foo"
        _, msg = Command(user_msg).commands[cmd]("foo", "-a")
        assert "You can only use -a in a DM" in msg

        # Test invalid query args
        _, msg = Command(user_msg).commands[cmd]("foo", "-i")
        assert 'the last keyword cannot be "-i"' in msg
        _, msg = Command(user_msg).commands[cmd]("foo", "-i", "bar")
        assert "The argument after '-i' must be an integer" in msg

        # Force a 500 status code
        m_request.return_value = create_mock_response(500, kind="gif")
        _, msg = Command(user_msg).commands[cmd]("foo")
        assert "I had trouble getting that query" in msg

        # Test no queries returned
        m_request.return_value = create_mock_response(200, kind="empty_gif")
        _, msg = Command(user_msg).commands[cmd]("foo")
        assert "there were no gifs for that query" in msg

        # Set to the discord abc to succeed with -a
        user_msg.channel = discord.abc.PrivateChannel()
        m_request.return_value = create_mock_response(200, kind="gif")
        type_, msg = Command(user_msg).commands[cmd]("foo", "-a")
        assert type_ == "list"
        assert isinstance(msg, list)

        # Get the first and second gifs explicitly from the mocked response
        m_request.return_value = create_mock_response(200, kind="gif")
        _, msg = Command(user_msg).commands[cmd]("-i", "0", "baz")
        assert msg == "foo"
        _, msg = Command(user_msg).commands[cmd]("-i", "1", "baz")
        assert msg == "bar"


@mock.patch("api.API._request")
def test_youtube(m_request):
    """ Test the youtube command """
    user_msg = create_user_msg()
    for cmd in {"!y", "!youtube"}:
        # Test no args throws an error
        _, msg = Command(user_msg).commands[cmd]("foo", "-i")
        assert 'the last keyword cannot be "-i"' in msg
        _, msg = Command(user_msg).commands[cmd]("foo", "-i", "bar")
        assert "The argument after '-i' must be an integer" in msg

        # Test with 500 status_code
        m_request.return_value = create_mock_response(500, kind="youtube")
        _, msg = Command(user_msg).commands[cmd]()
        assert "I had trouble getting that query" in msg

        # Test no videos in response
        m_request.return_value = create_mock_response(200, kind="empty_youtube")
        _, msg = Command(user_msg).commands[cmd]()
        assert "there were no videos for that query" in msg

        # Get the first and second videos explicitly from the mocked response
        m_request.return_value = create_mock_response(200, kind="youtube")
        _, msg = Command(user_msg).commands[cmd]("-i", "0", "baz")
        assert msg == "https://www.youtube.com/watch?v=foo"
        _, msg = Command(user_msg).commands[cmd]("-i", "1", "baz")
        assert msg == "https://www.youtube.com/watch?v=bar"


@mock.patch("commands.requests.get")
def test_waifu(m_request):
    """ Test the waifu command """
    user_msg = create_user_msg()
    for cmd in {"!w", "!waifu"}:
        # Test non-200 status code
        # We don't care about content since it's mocked. Just reuse the youtube one
        m_request.return_value = create_mock_response(500, kind="youtube")
        _, msg = Command(user_msg).commands[cmd]()
        assert "I coudn't get that waifu" in msg

        # Test 200 status code
        # We don't care about content since it's mocked. Just reuse the youtube one
        m_request.return_value = create_mock_response(200, kind="youtube")
        kind, msg = Command(user_msg).commands[cmd]()
        assert kind == "file"
        assert msg == "temp.jpg"


def test_nut():
    """ Test the nut command """
    user_msg = create_user_msg()
    for cmd in {"!n", "!nut"}:
        _, msg = Command(user_msg).commands[cmd]()
        assert "Remember" in msg
        assert ", don't " in msg
