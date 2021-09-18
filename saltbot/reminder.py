"""
    Reminder class for handing reminders
"""
import asyncio
from datetime import datetime
import json
import glob
import os
from pathlib import Path
import time

import pytz

from timelength import TimeLength

REMINDER_DIR = os.path.join(str(Path.home()), ".config/saltbot/reminders")
os.makedirs(REMINDER_DIR, exist_ok=True)

REMINDER_HELP_MSG = (
    '```Set a reminder, show reminders or delete a reminder.\n\nTo set one:\n"!remind set finish '
    'fixing saltbot bugs in 4 hours"\n\nTo show all reminders:\n"!remind show"\n\nTo delete:'
    '\n"!remind delete <ID>"\n\nTo delete a reminder, you need to know the ID. The "!remind show "'
    "command lists all your reminders and IDs```"
)


class ReminderError(Exception):
    """ Raised when there is an issue with a reminder operation """


class ReminderFile(dict):
    """ Reminder File Object to Read and Write from memory """

    def __init__(self, **kwargs):
        super().__init__(**kwargs)

        path = kwargs.pop("path")

        self.msg = kwargs.get("msg")
        self.unique_id = kwargs.get("unique_id")
        self.timeout = kwargs.get("timeout")
        self._channel = kwargs.get("channel")
        self._path = f"{path}/{self.unique_id}.json"

        if os.path.exists(self._path):
            self._read()

    def __str__(self):
        dt_obj = datetime.fromtimestamp(self.timeout)
        timezone = pytz.timezone("US/Eastern")
        timezone.localize(dt_obj)

        formatted_time = dt_obj.strftime("%b %d, %Y at %I:%M:%S %p ET")

        return f'{self.unique_id}: "{self.msg}" at {formatted_time}'

    @property
    def channel(self):
        """ Read only property to get the channel ID """
        return self._channel

    @property
    def expired(self):
        """ Property to determine if a reminder has expired """
        return self.timeout < time.time()

    def write(self):
        """ Write a reminder to disk """
        with open(self._path, "w") as stream:
            stream.write(json.dumps(self))

    def delete(self):
        """ Delete a reminder from disk """
        try:
            os.remove(self._path)
        except FileNotFoundError as error:
            raise ReminderError(
                "```That reminder ID does not exist! :o\n"
                'Type "!remind show" to see all your reminders```'
            ) from error

    def _read(self):
        with open(self._path) as stream:
            data = json.load(stream)

        self.msg = data["msg"]
        self.unique_id = data["unique_id"]
        self.timeout = data["timeout"]
        self._channel = data["channel"]


class ReminderFileHandler:
    """ File Finder/Handler """

    def __init__(self, user):
        self._user = user
        self.path = f"{REMINDER_DIR}/{user}"
        os.makedirs(self.path, exist_ok=True)

    def load(self, reminder_id):
        """ Load an existing reminder by id """
        return ReminderFile(path=self.path, unique_id=reminder_id)

    def write(self, **kwargs):
        """ Write a reminder to disk and return the object """
        kwargs["path"] = self.path
        reminder_file = ReminderFile(**kwargs)
        reminder_file.write()
        return str(reminder_file)

    @property
    def reminders(self):
        """ Get all reminders for a specified user """
        reminders = []
        files = glob.glob(f"{self.path}/*")
        for file_ in files:
            unique_id = get_id_from_full_path(file_)
            reminders.append(ReminderFile(path=self.path, unique_id=unique_id))

        return reminders


class Reminder:
    """ Reminder class """

    def __init__(self, user, channel, *args):
        # Convert the tuple to a list
        self._args = list(args)

        self._user = user
        self._channel = channel
        self._cmd = self._args.pop(0)

    def execute(self):
        """ Parse through the reminder and set appropriate instance vars """
        if self._cmd.lower() == "set":
            try:
                unit = self._args.pop(-1)
                amount_of_time = self._args.pop(-1)

                # Remove "in"
                self._args.remove("in")

                return self.set_reminder(unit, amount_of_time, msg=" ".join(self._args))

            except (IndexError, ValueError) as error:
                raise ReminderError(
                    "```Reminder time must be in format <amount> <unit> (ex 5 minutes)```"
                ) from error

        if self._cmd.lower() == "show":
            return self.show_reminders()

        if self._cmd.lower() == "delete":
            reminder_id = self._args.pop(-1)
            return self.delete(reminder_id)

        raise ReminderError(
            f"```Invalid command {self._cmd}\nValid commands: ('set', 'show', 'delete')```"
        )

    def set_reminder(self, unit, amount_of_time, msg):
        """ Write a reminder to disk """
        try:
            time_length = TimeLength(unit, amount_of_time)
        except ValueError as error:
            raise ReminderError(str(error)) from error

        reminder_data = {
            "msg": msg,
            "unique_id": time_length.unique_id,
            "timeout": time_length.timeout,
            "channel": self._channel,
        }

        handler = ReminderFileHandler(self._user)
        reminder = handler.write(**reminder_data)

        return f"```{reminder}```"

    def show_reminders(self):
        """ Show all reminders """
        reminder_str = "```Your Reminders:\n"
        for reminder in ReminderFileHandler(self._user).reminders:
            reminder_str += f"\n{reminder}"

        return f"{reminder_str}```"

    def delete(self, reminder_id):
        """ Delete a user's remind using the id """
        reminder = ReminderFileHandler(self._user).load(reminder_id)
        reminder.delete()

        return f"```Ok. I've deleted:\n{reminder}```"


async def monitor_reminders(discord_client):
    """ Check reminder files for expiry every 5 seconds """
    while True:
        for user_path in glob.glob(f"{REMINDER_DIR}/*"):
            user = user_path.split("/")[-1]
            handler = ReminderFileHandler(user)

            for reminder_file in glob.glob(f"{user_path}/*"):
                reminder_id = get_id_from_full_path(reminder_file)
                reminder = handler.load(reminder_id)
                if reminder.expired:
                    channel = discord_client.get_channel(reminder.channel)
                    await channel.send(f"```Remember:\n{reminder.msg}```")
                    try:
                        reminder.delete()
                    except ReminderError:
                        # If the reminder somehow gets deleted, ignore it
                        pass

        await asyncio.sleep(5)


def get_id_from_full_path(path):
    """ Pull the id from the path name """
    return os.path.splitext(path)[0].split("/")[-1]
