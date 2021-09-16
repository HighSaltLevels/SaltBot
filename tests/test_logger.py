""" Logging Test Module """

import os
import sys
from unittest import mock

import pytest

from logger import Logger


class MockPrint:
    """
    Mock print object that raises exception the first time, but
    acts normally the rest of the time
    """

    # We need to mock out print, so we won't be using all of print's args.
    # This also means we don't really need any public methods
    # pylint: disable=unused-argument,too-few-public-methods
    def __init__(self, msg="", **kwargs):
        self._raise_exception = True

    def __call__(self, msg="", **kwargs):
        if self._raise_exception:
            self._raise_exception = False
            raise UnicodeDecodeError("1", b"\x0a", 3, 4, "5")

        sys.stdout.write(msg)


@pytest.fixture(name="log")
def create_logger():
    """ Create a logger object """
    yield Logger()
    os.remove("log.txt")


def test_logger(log, capfd):
    """ Test logger operations """
    assert os.path.isfile("log.txt")

    # Test logging normal text
    log.log("foo")
    assert "foo" in capfd.readouterr().out

    # Test with unicode characters
    with mock.patch("builtins.print", new_callable=MockPrint):
        log.log("foo")
    assert "had unicode bytes" in capfd.readouterr().out
