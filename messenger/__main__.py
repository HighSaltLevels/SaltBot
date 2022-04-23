""" SaltBot Messenger """

import os

from controller import CLIENT

BOT_TOKEN = os.getenv("BOT_TOKEN")

if __name__ == "__main__":
    CLIENT.run(BOT_TOKEN)
