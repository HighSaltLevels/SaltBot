""" TimeLength Test Module """

import time

import pytest

from timelength import TimeLength, UNIT_DICT


def test_different_ids():
    """Test timelengths have different IDs  10 times"""
    time_length_id = TimeLength("second", 1).unique_id
    for _ in range(10):
        assert time_length_id != TimeLength("second", 1).unique_id
