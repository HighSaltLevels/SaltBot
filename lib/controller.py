import discord

from lib.commands import Command
from lib.logger import Logger

LOGGER = Logger()
CLIENT = discord.Client()


@CLIENT.event
async def on_message(msg):
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

        else:
            type_, resp = bot_cmd.commands[cmd](*args)
            LOGGER.log_sent(msg.author, msg.channel, cmd)

            if type_ == "text":
                await msg.channel.send(resp)
            elif type_ == "file":
                await msg.channel.send(file=discord.File(resp))
            elif type_ == "list":
                for item in resp:
                    await msg.channel.send(item)
            else:
                err_msg = "```Unexpected error :(```"
                await msg.channel.send(err_msg)


@CLIENT.event
async def on_ready():
    LOGGER.log("Logged in as")
    LOGGER.log(CLIENT.user.name)
    LOGGER.log(str(CLIENT.user.id))
    await CLIENT.change_presence(activity=discord.Game(name="The Salt Shaker"))
