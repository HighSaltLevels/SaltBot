""" TimeLength Test Module """

import time

import pytest

from timelength import TimeLength, UNIT_DICT

def test_timelength():
    """ Test timelength functions work like they should """
    # Test unit in dict
    for unit in UNIT_DICT:
        time_length = TimeLength(unit, 5)

    # Test unit not in dict
    with pytest.raises(ValueError) as error:
        time_length = TimeLength("foo", 5)

    # Test Timeout Calculation
    time_length = TimeLength("hours", 5)
    assert time_length.timeout <= UNIT_DICT["hours"] * 5 + time.time()

    assert "Invalid unit foo" in str(error)

def test_different_ids():
    """ Test timelengths have different IDs  10 times """
    time_length_id = TimeLength("second", 1).unique_id
    for _ in range(10):
        assert time_length_id != TimeLength("second", 1).unique_id
