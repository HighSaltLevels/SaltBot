""" Controller module for handling coroutines """
import traceback

import discord

from commands import Command
from logger import Logger
from poll import monitor_polls

LOGGER = Logger()
CLIENT = discord.Client()


@CLIENT.event
async def on_message(msg):
    """ Main Message Event Handler """
    # Only do something if command starts with ! or bot is not sending message
    if msg.author != CLIENT.user and msg.content.startswith("!"):
        LOGGER.log_received(msg.author, msg.channel, msg.content)

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
            LOGGER.log_sent(msg.author, msg.channel, cmd)

            if type_ == "text":
                await msg.channel.send(resp)
            elif type_ == "file":
                await msg.channel.send(file=discord.File(resp))
            elif type_ == "list":
                for item in resp:
                    await msg.channel.send(item)
            elif type_ == "user":
                await msg.author.send(resp)

        except Exception:  #  pylint: disable=broad-except
            traceback.print_exc()
            await msg.channel.send("```Unexpected error :(```")


@CLIENT.event
async def on_ready():
    """ Print out basic info and set status on startup """
    LOGGER.log("Logged in as")
    LOGGER.log(CLIENT.user.name)
    LOGGER.log(str(CLIENT.user.id))
    await CLIENT.change_presence(activity=discord.Game(name="The Salt Shaker"))
    CLIENT.loop.create_task(monitor_polls(CLIENT))
