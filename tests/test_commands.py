""" Command Test Module """

from unittest import mock

import discord
import pytest

from commands import Command, MSG_DICT
from poll import Poll, POLL_HELP_MSG
from timelength import TimeLength
from reminder import REMINDER_HELP_MSG
from .utils import create_user_msg, create_mock_response


def test_help():
    """Test the help command"""
    user_msg = create_user_msg()
    # Test all commands are documented in help
    for _help in {"!h", "!help"}:
        _, msg = Command(user_msg).commands[_help]()
        for key, value in MSG_DICT.items():
            assert key in msg
            assert value in msg


def test_jeopardy():
    """Test the jeopardy command"""
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
    """Test the whisper command"""
    user_msg = create_user_msg()
    for cmd in {"!pm", "!whisper"}:
        _, msg = Command(user_msg).commands[cmd]()
        assert "Hello foo" in msg


@mock.patch("api.API._request")
def test_gif(m_request):
    """Test the gif command"""
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
    """Test the youtube command"""
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
    """Test the waifu command"""
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
    """Test the nut command"""
    user_msg = create_user_msg()
    for cmd in {"!n", "!nut"}:
        _, msg = Command(user_msg).commands[cmd]()
        assert "Remember" in msg
        assert ", don't " in msg
