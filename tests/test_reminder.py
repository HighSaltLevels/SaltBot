""" Reminder Test Module """

from datetime import datetime
import os
import shutil
import time

import pytz
import pytest

from reminder import (
    Reminder,
    ReminderFile,
    ReminderError,
    ReminderFileHandler,
    REMINDER_DIR,
)


@pytest.fixture(name="reminder_info")
def create_reminder_info():
    """ Create fake reminder data """
    # Set up reminder path for test
    path = f"{REMINDER_DIR}/foo"
    os.makedirs(path, exist_ok=True)

    yield {
        "path": path,
        "unique_id": "abcdefghijklmnop",
        "msg": "hi",
        "timeout": 420,
        "channel": 32,
    }

    shutil.rmtree(path, ignore_errors=True)


@pytest.fixture(name="file_handler")
def create_file_handler():
    """ Create a ReminderFileHandler object and clean up """
    yield ReminderFileHandler(user="foo")
    shutil.rmtree(f"{REMINDER_DIR}/foo", ignore_errors=True)


def test_reminder_file(reminder_info):
    """ Test reminder file operations """

    # Test with a reminder file that doesn't exist
    reminder = ReminderFile(**reminder_info)
    path = f'{reminder_info["path"]}/{reminder_info["unique_id"]}.json'
    assert not os.path.isfile(path)

    assert reminder.channel == reminder_info["channel"]
    assert reminder.expired <= reminder_info["timeout"] + time.time()

    dt_obj = datetime.fromtimestamp(reminder_info["timeout"])
    pytz.timezone("US/Eastern").localize(dt_obj)
    formatted_time = dt_obj.strftime("%b %d, %Y at %I:%M:%S %p ET")
    assert formatted_time in str(reminder)

    # Test writing to disk
    reminder.write()
    assert os.path.isfile(path)

    # Test deleting a reminder
    reminder.delete()
    assert not os.path.isfile(path)

    # Test deleting a reminder that does not exist
    with pytest.raises(ReminderError) as error:
        reminder.delete()

    assert "That reminder ID does not exist" in str(error)


def test_load_existing_reminder(reminder_info):
    """ Test loading an existing reminder from disk """

    # Create a reminder to load
    reminder = ReminderFile(**reminder_info)
    path = f'{reminder_info["path"]}/{reminder_info["unique_id"]}.json'
    assert not os.path.isfile(path)

    reminder.write()
    assert os.path.isfile(path)

    # Test loading a saved reminder
    loaded_reminder = ReminderFile(
        path=reminder_info["path"], unique_id=reminder_info["unique_id"]
    )
    assert reminder.unique_id == loaded_reminder.unique_id
    loaded_reminder.delete()


def test_reminder_filehandler(file_handler, reminder_info):
    """ Test operations on filehandler """
    unique_id = reminder_info["unique_id"]

    assert len(file_handler.reminders) == 0, "FileHandler was not empty"
    assert unique_id in file_handler.write(**reminder_info)
    assert len(file_handler.reminders) == 1, "FileHandler did not have 1 reminder"
    assert (
        unique_id == file_handler.load(unique_id).unique_id
    ), "FileHandler did not return the right file"


def test_invalid_reminder():
    """ Test invalid reminder throws error """
    # Test invalid command
    with pytest.raises(ReminderError) as error:
        Reminder("user", "channel", "foo").execute()
    assert "Invalid command foo" in str(error)

    # Test invalid unit
    args = ["set", "foo", "in", "5", "bar"]
    with pytest.raises(ReminderError) as error:
        Reminder("user", "channel", *args).execute()
    assert "Invalid unit bar" in str(error)

    # Test invalid time spec
    args = ["set", "foo", "in", "bar"]
    with pytest.raises(ReminderError) as error:
        Reminder("user", "channel", *args).execute()
    assert "Reminder time must be in format" in str(error)


def test_reminder_operations():
    """ Test reminder operations """

    # Test setting the reminder
    args = ["set", "foo", "in", "4", "seconds"]
    reminder = Reminder("user", "channel", *args).execute()
    assert ': "foo" at' in reminder

    # Test showing reminders
    assert "Your Reminders" in Reminder("user", "channel", "show").execute()

    # Test deleting reminder
    id_ = reminder.replace("```", "").split(":")[0]
    assert "Ok. I've deleted" in Reminder("user", "channel", "delete", id_).execute()
