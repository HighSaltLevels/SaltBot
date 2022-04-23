""" Controller Test Module """

from unittest import mock

import pytest

from controller import on_message, on_ready

pytestmark = pytest.mark.asyncio


@pytest.fixture(name="m_client")
async def create_mock_client():
    """Create a mocked discord client"""
    with mock.patch("controller.CLIENT") as m_client:
        m_client.user = "different-user"
        yield m_client


@pytest.fixture(name="m_msg")
async def create_mock_msg():
    """Create a mocked discord msg object"""
    attrs = {
        "channel.send.return_value": None,
        "author.send.return_value": None,
    }
    m_msg = mock.AsyncMock()
    m_msg.configure_mock(**attrs)
    m_msg.author = "user"
    m_msg.content = "!help"
    yield m_msg


@pytest.fixture(name="m_help")
async def create_mock_help():
    """Create a mock help command object"""
    with mock.patch("commands.Command.help") as m_help:
        yield m_help


async def test_on_message_from_saltbot(m_client, m_msg):
    """Test the response handler for on_message"""
    m_msg.author = "user"
    m_client.user = "user"
    await on_message(m_msg)
    m_msg.channel.send.assert_not_called()


async def test_unknown_command(m_msg):
    """Test unknown commands return instructions for help"""
    m_msg.content = "!invalid command"
    await on_message(m_msg)
    m_msg.channel.send.assert_called_with(
        "```Hello. I'm sorry I don't understand !invalid. Please type "
        '"!help" to see a list of available commands\n```'
    )


async def test_command_response_types(m_help, m_msg):
    """Test the different response types (text, file, list, etc...)"""
    # Test a text
    m_help.return_value = "text", "foo"
    await on_message(m_msg)
    m_msg.channel.send.assert_called_with("foo")

    # Test a file
    m_msg.reset_mock()
    m_help.return_value = "file", "foo"
    await on_message(m_msg)
    m_msg.channel.send.assert_called()

    # Test a list
    m_msg.reset_mock()
    m_help.return_value = "list", range(5)
    await on_message(m_msg)
    m_msg.channel.send.assert_has_calls(mock.call(num) for num in range(5))

    # Test a user
    m_msg.reset_mock()
    m_msg.author = mock.AsyncMock()
    m_help.return_value = "user", "foo"
    await on_message(m_msg)
    m_msg.author.send.assert_called_with("foo")
    m_msg.channel.send.assert_not_called()


async def test_command_exception(m_help, m_msg, capfd):
    """Test error response is returned to user when there's an unexpected error"""
    m_help.side_effect = Exception("foo")
    await on_message(m_msg)
    output = capfd.readouterr().out
    assert "foo" in output
    assert "Unexpected error with id" in output
    m_msg.channel.send.assert_called_once()
