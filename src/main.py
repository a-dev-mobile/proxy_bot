import logging
from telegram import Update
from telegram import ReplyKeyboardMarkup, KeyboardButton
from telegram.ext import ApplicationBuilder, ContextTypes, CommandHandler
from telegram.ext import  CommandHandler
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

list_proxy = []
rand = False



async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    msg = update.message
    chat = update.effective_chat
    if msg and chat and msg.text:
        keyboard = [
            [
                KeyboardButton("1 вар."),
                KeyboardButton("2 вар."),
            ],
        ]
        reply_markup = ReplyKeyboardMarkup(keyboard, resize_keyboard=True)
        await msg.reply_text("Выберите кнопку для поиска работающих IP прокси:", reply_markup=reply_markup)


# Обработчик нажатий на кнопки
async def button_click(update, context):
    text = update.message.text
    print("\n--button_click--")
    print(text)
    global is_running
    is_running = True
    if text == "1 вар.":
        global rand
        rand = True
        list_proxy.clear()
        await start_proxy(update, context)
    elif text == "2 вар.":
        # list_proxy.clear()
        rand = False
        await start_proxy(update, context)






async def start_proxy(update: Update, context: ContextTypes.DEFAULT_TYPE):
    msg = update.message
    chat = update.effective_chat
    if msg and chat and msg.text:
        for i in range(10): 

            await context.bot.send_message(
                chat_id=chat.id,
                text='... '+str(i+1)+'/10 ...',
            )
            print("\n--search proxy--")
            print("Random " + str(rand))
            print("list " + str(list_proxy))
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
                    await send_text(context, chat, "--Нашли--")
                    break

                else:
                    await send_text(context, chat, format_proxy)
                    await send_text(context, chat, formatted_json)
        

if __name__ == "__main__":
    application = (
        ApplicationBuilder()
        .token("6070162821:AAEj_TzX1D0SrcFrcXeEnwBSny5C4FxxBoc")
        .build()
    )

    button_click_handler = MessageHandler(filters.TEXT & (~filters.COMMAND), button_click)
    start_handler = CommandHandler("start", start)
    start_proxy_handler = CommandHandler("start_proxy", start_proxy)

    application.add_handler(start_handler)
    application.add_handler(button_click_handler)
    application.add_handler(start_proxy_handler)

    application.run_polling()
