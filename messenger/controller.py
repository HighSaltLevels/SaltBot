""" Controller Entrypoint Module """

import asyncio
import time

import discord

from common.logger import LOGGER
from common.k8s.configmap import ConfigMap
from messenger import Messenger

CLIENT = discord.Client()


class Controller:
    """Controller class"""

    def __init__(self, client):
        self._messenger = Messenger(client)

        self._polls = []
        self._reminders = []

    async def start(self):
        """Start waiting for polls/reminders to expire"""
        LOGGER.log("Started up messenger to monitor for polls/reminders")
        while True:
            self.get_all_config_maps()
            await self._check_expiry()
            await asyncio.sleep(5)

    def get_all_config_maps(self):
        """List all config_maps and sort them into separate buckets"""
        self._polls = []
        self._reminders = []

        config_maps = ConfigMap().list()
        for config_map in config_maps:
            kind = config_map.get("kind", "")
            if kind == "poll":
                self._polls.append(config_map)

            elif kind == "reminder":
                self._reminders.append(config_map)

    async def _check_expiry(self):
        """Check if any polls or reminders are ready to send and send them"""
        now = time.time()
        for poll in self._polls:
            expiry = float(poll.get("expiry", 0.0))
            if expiry < now:
                name = f"poll-{poll['unique_id']}"
                self._delete_config_map(name)
                self._polls.remove(poll)
                await self._messenger.send_poll_results(poll)

        for reminder in self._reminders:
            expiry = float(reminder.get("expiry", 0.0))
            if expiry < now:
                name = f"reminder-{reminder['unique_id']}"
                msg = f"```Remember:\n{reminder['msg']}```"
                self._delete_config_map(name)
                self._reminders.remove(reminder)
                await self._messenger.send(msg, reminder)

    @staticmethod
    def _delete_config_map(name):
        """Delete the poll config map"""
        LOGGER.log(f"Deleting ConfigMap {name}")
        ConfigMap().delete(name)
        LOGGER.log(f"Successfully deleted ConfigMap {name}")


@CLIENT.event
async def on_ready():
    """Log in as the saltbot user"""
    LOGGER.log("Logged in as")
    LOGGER.log(CLIENT.user.name)
    LOGGER.log(CLIENT.user.id)
    controller = Controller(CLIENT)
    CLIENT.loop.create_task(controller.start())
