import logging
from telegram import Update
from telegram.ext import ApplicationBuilder, ContextTypes, CommandHandler
import json
from telegram.ext import (
    filters,
    MessageHandler,
    ApplicationBuilder,
    CommandHandler,
    ContextTypes,
)
from fp.fp import FreeProxy
from utils import parse_ip_port, send_text
from proxy_checking import ProxyChecker

logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO
)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    msg = update.message
    chat = update.effective_chat
    if msg and chat and msg.text:
        await context.bot.send_message(
            chat_id=chat.id,
            text='''
"Я готов оказать вам помощь в различных вопросах и задачах. 
Вы можете взаимодействовать со мной, отправляя следующие команды:

/start_proxy - для поиска свободных прокси-серверов
/start_random_proxy  - для случайного поиска прокси-сервера
/reset_proxy - для сброса результатов поиска прокси-серверов''',
        )


# 1 сравнивать по статусу   "status": true,   "status": false
list_proxy = []
rand = False


async def start_random_proxy(update: Update, context: ContextTypes.DEFAULT_TYPE):
    msg = update.message
    chat = update.effective_chat
    if msg and chat and msg.text:
        global rand
        rand = True
        await context.bot.send_message(
            chat_id=chat.id,
            text='''
Включение поиска рандомных прокси
''',
        )
        await start_proxy(update, context )


async def reset_proxy(update: Update, context: ContextTypes.DEFAULT_TYPE, ):
    msg = update.message
    chat = update.effective_chat
    if msg and chat and msg.text:
        list_proxy.clear()
        global rand
        rand = False
        await context.bot.send_message(
            chat_id=chat.id,
            text='''
Очистка сохраненных прокси 
Выключения поиска рандомных прокси
''',
        )
        await start_proxy(update, context)


async def start_proxy(update: Update, context: ContextTypes.DEFAULT_TYPE):
    msg = update.message
    chat = update.effective_chat
    if msg and chat and msg.text:

        while True:
            await context.bot.send_message(
            chat_id=chat.id,
            text="...",
        )
            print("\n--search proxy--")
            print("Random " +str(rand))
            print("list " +str(list_proxy))
            
            proxy = FreeProxy(rand=rand).get()
            ip, port = parse_ip_port(proxy)
            format_proxy = str(ip) + ":" + str(port)

            if list_proxy:
                found_in_list = False
                for item in list_proxy:
                    if format_proxy in item:
                        print("Текст найден в списке:", item)
                        found_in_list = True
                        

                if not found_in_list:
                    list_proxy.append(format_proxy)
                if found_in_list:
                    continue
            else:
                list_proxy.append(format_proxy)

            try:
                checker = ProxyChecker()
            except Exception as e:
                print("Ошибка при создании экземпляра ProxyChecker:", e)
                continue
            valid = checker.check_proxy(format_proxy)
            formatted_json = json.dumps(valid, indent=4)
            print(type(valid))

            if isinstance(valid, dict):
                if valid.get("status") == True:
                    await send_text(context, chat, format_proxy)
                    await send_text(context, chat, formatted_json)
                    await send_text(context, chat, '--ok--')
                    break
                else:
                    await send_text(context, chat, format_proxy)
                    await send_text(context, chat, formatted_json)


async def echo(update: Update, context: ContextTypes.DEFAULT_TYPE):
    msg = update.message
    chat = update.effective_chat
    if msg and chat and msg.text:
        await context.bot.send_message(chat_id=chat.id, text=msg.text)
    else:
        print("No message found in the update object")


if __name__ == "__main__":
    application = (
        ApplicationBuilder()
        .token("6070162821:AAEj_TzX1D0SrcFrcXeEnwBSny5C4FxxBoc")
        .build()
    )

    echo_handler = MessageHandler(filters.TEXT & (~filters.COMMAND), echo)
    start_handler = CommandHandler("start", start)
    start_proxy_handler = CommandHandler("start_proxy", start_proxy)
    reset_proxy_handler = CommandHandler("reset_proxy", reset_proxy)
    start_random_proxy_handler = CommandHandler("start_random_proxy", start_random_proxy)

    application.add_handler(start_handler)
    application.add_handler(start_proxy_handler)
    application.add_handler(reset_proxy_handler)
    application.add_handler(start_random_proxy_handler)

    application.run_polling()
