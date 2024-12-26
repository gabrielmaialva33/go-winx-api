# 🌟 GO Winx Session String Generator 🌟
import asyncio
import os
import sys

# Ensure the environment supports UTF-8
os.environ["PYTHONIOENCODING"] = "UTF-8"

from pyrogram import Client
from telethon import TelegramClient
from telethon.sessions import StringSession

print("🎉 Welcome to GO Winx Session String Generator! 🎉")

APP_ID = input("🔵 Enter your App ID: ")
API_HASH = input("🔴 Enter your API Hash: ")

SESSION_MSG = """
**🌟 GO Winx Session Generator 🌟**

📢 Hello! You have successfully generated the {type} session string for GO Winx Userbot as follows:

`{session}`

⚠️ **Note**: **DO NOT SHARE** this session string with anyone as it may cause hijacking of your account.
"""

SESSION_TYPE = input(
    """🌌 Please specify session type:\n    🔢 Enter 0 for Telethon\n    🔢 Enter 1 for Pyrogram\n    👇 Your answer: """)


async def generate_telethon_session(app_id, api_hash):
    async with TelegramClient(StringSession(), app_id, api_hash) as c:
        session_string = c.session.save()
        await c.send_message("me", SESSION_MSG.format(type="telethon", session=session_string))
        print("🔗 Telethon session string sent to your saved messages! 🏠")


if SESSION_TYPE == "0":
    asyncio.run(generate_telethon_session(APP_ID, API_HASH))
elif SESSION_TYPE == "1":
    client = Client("", APP_ID, API_HASH, in_memory=True)
    client.start()
    client.send_message("me", SESSION_MSG.format(type="pyrogram", session=client.export_session_string()))
    print("🔗 Pyrogram session string sent to your saved messages! 🏠")
else:
    print("❌ Your input was invalid. Expected 0 or 1. Please try again! ⚠️")
    sys.exit(1)

print("\n🎉 You successfully generated the session string for GO Winx Userbot! 🎉")
print("🔍 Check your saved messages on Telegram to retrieve it. 📢")
