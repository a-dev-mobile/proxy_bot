import re

from telegram import Chat
from telegram.ext import ContextTypes


def parse_ip_port(url):
    match = re.match(r"http://([\d.]+):(\d+)", url)
    if match:
        ip = match.group(1)
        port = match.group(2)
        return ip, int(port)
    else:
        return None, None


async def send_text(context: ContextTypes.DEFAULT_TYPE, chat: Chat, ip):
    await context.bot.send_message(
        chat_id=chat.id,
        text=str(ip),
    )
