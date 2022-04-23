"""
    TimeLength module for operations on units like "3 seconds"
"""

import time
import uuid

UNIT_DICT = {
    "year": 31536000,
    "years": 31536000,
    "month": 2592000,
    "months": 2592000,
    "weeks": 604800,
    "week": 604800,
    "days": 86400,
    "day": 86400,
    "hours": 3600,
    "hour": 3600,
    "minutes": 60,
    "minute": 60,
    "seconds": 1,
    "second": 1,
}


class TimeLength:
    """TimeLength class"""

    def __init__(self, unit, amount_of_time):
        try:
            multiplier = UNIT_DICT[unit]
        except KeyError as error:
            raise ValueError(f"```Invalid unit {unit}```") from error

        self._id = str(uuid.uuid4()).split("-")[0]
        self._timeout = str(int(time.time()) + int(amount_of_time) * multiplier)

    @property
    def timeout(self):
        """Epoch timeout"""
        return self._timeout

    @property
    def unique_id(self):
        """Unique ID"""
        return self._id
