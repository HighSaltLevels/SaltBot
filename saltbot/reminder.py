"""
    Reminder class for handing reminders
"""
from datetime import datetime
from http import HTTPStatus
import uuid

from kubernetes.client.rest import ApiException
import pytz

from timelength import TimeLength
from common.k8s.configmap import ConfigMap

REMINDER_HELP_MSG = (
    '```Set a reminder, show reminders or delete a reminder.\n\nTo set one:\n"!remind set finish '
    'fixing saltbot bugs in 4 hours"\n\nTo show all reminders:\n"!remind show"\n\nTo delete:'
    '\n"!remind delete <ID>"\n\nTo delete a reminder, you need to know the ID. The "!remind show "'
    "command lists all your reminders and IDs```"
)


class ReminderError(Exception):
    """Raised when there is an issue with a reminder operation"""


class Reminder:
    """Reminder class"""

    def __init__(self, user, channel, cmd, author, *args):
        # Convert the tuple to a list
        self._args = list(args)

        # Convert the channel to a string so it can be put in a configmap
        self._channel = str(channel)
        self._user = user
        self._cmd = cmd
        self._author = author

    def execute(self):
        """Parse through the reminder and set appropriate instance vars"""
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
            # Since labels cannot use "#", replace with a "."
            selector = f"user={self._user}".replace("#", ".")
            return self.list(selector=selector)

        if self._cmd.lower() == "delete":
            reminder_id = self._args.pop(-1)
            return self.delete(reminder_id)

        raise ReminderError(
            f"```Invalid command {self._cmd}\nValid commands: ('set', 'show', 'delete')```"
        )

    def set_reminder(self, unit, amount_of_time, msg):
        """Write a reminder to disk"""
        # Generate an ID using the last part of a UUID4 string
        _id = str(uuid.uuid4()).split("-")[-1]
        try:
            time_length = TimeLength(unit, amount_of_time)
        except ValueError as error:
            raise ReminderError(str(error)) from error

        name = f"reminder-{_id}"
        # Since labels cannot use "#", replace with a "."
        labels = {"user": self._user.replace("#", ".")}
        body = {
            "kind": "reminder",
            "claimed": "false",
            "author": str(self._author.id),
            "msg": msg,
            "unique_id": _id,
            "expiry": time_length.timeout,
            "channel": self._channel,
        }

        config_map = ConfigMap()
        config_map.create(name=name, labels=labels, data=body)

        return f"```Reminder set with ID {_id}```"

    def list(self, selector):
        """Show all reminders"""
        reminders = ConfigMap().list(label_selector=selector)

        final_resp = "```Your Reminders:\n"
        for reminder in reminders:
            reminder_str = self._get_reminder_str(reminder)
            final_resp += f"\n{reminder_str}"

        return f"{final_resp}```"

    @staticmethod
    def delete(reminder_id):
        """Delete a user's remind using the id"""
        name = f"reminder-{reminder_id}"
        try:
            ConfigMap().delete(name)
        except ApiException as error:
            if error.status == HTTPStatus.NOT_FOUND:
                raise ReminderError(
                    f"```Reminder {reminder_id} does not exist! "
                    'Check reminders using "!remind show"```'
                ) from error

            raise error

        return f"```Ok. I've deleted reminder {reminder_id}```"

    @staticmethod
    def _get_reminder_str(reminder):
        """Take the raw reminder response and convert it to human readable"""
        # ConfigMaps only store strings so we need to parse to int
        dt_obj = datetime.fromtimestamp(int(reminder["expiry"]))
        timezone = pytz.timezone("US/Eastern")
        timezone.localize(dt_obj)

        formatted_time = dt_obj.strftime("%b %d, %Y at %I:%M:%S %p ET")

        return f"""{reminder['unique_id']}: "{reminder['msg']}" at {formatted_time}"""
