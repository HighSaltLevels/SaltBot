""" Controller module for handling coroutines """
import uuid
import traceback

import discord

from commands import Command
from logger import Logger
from poll import monitor_polls
from reminder import monitor_reminders

LOGGER = Logger()
CLIENT = discord.Client()


@CLIENT.event
async def on_message(msg):
    """ Main Message Event Handler """
    # Only do something if command starts with ! or bot is not sending message
    if msg.author != CLIENT.user and msg.content.startswith("!"):
        LOGGER.log(f"Got {msg.content} from {msg.author} in {msg.channel}")

        args = msg.content.split(" ")
        cmd = args.pop(0)
        bot_cmd = Command(msg)

        if cmd not in bot_cmd.commands:
            await msg.channel.send(
                f"```Hello. I'm sorry I don't understand {cmd}. Please type "
                '"!help" to see a list of available commands\n```'
            )
            return

        try:
            type_, resp = bot_cmd.commands[cmd](*args)

            if type_ == "text":
                await msg.channel.send(resp)
            elif type_ == "file":
                await msg.channel.send(file=discord.File(resp))
            elif type_ == "list":
                for item in resp:
                    await msg.channel.send(item)
            elif type_ == "user":
                await msg.author.send(resp)

        except Exception as error:  #  pylint: disable=broad-except
            error_msg = f"Unexpected error with id: {uuid.uuid4()}"
            print(f"{error_msg} {error}")
            traceback.print_exc()
            await msg.channel.send(f"```{error_msg} :(```")


@CLIENT.event
async def on_ready():
    """ Print out basic info and set status on startup """
    LOGGER.log("Logged in as")
    LOGGER.log(CLIENT.user.name)
    LOGGER.log(str(CLIENT.user.id))
    await CLIENT.change_presence(activity=discord.Game(name="The Salt Shaker"))

    for coroutine in (monitor_polls, monitor_reminders):
        CLIENT.loop.create_task(coroutine(CLIENT))
