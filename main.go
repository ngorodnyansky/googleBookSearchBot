package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func startCommand(bot *tgbotapi.BotAPI, chatId int64) {
	startMessage := "Введите текст по которому необходимо сделать поиск в Google Books:"
	sendMessage(bot, chatId, startMessage)
}

func helpCommand(bot *tgbotapi.BotAPI, chatId int64) {
	helpMessage := "Этот бот делает запрос по введённому тексту в Goole книги и возвращает полученный результат. Присто напишите текст по которому хотите сделать поиск."
	sendMessage(bot, chatId, helpMessage)
}

func defaultCommand(bot *tgbotapi.BotAPI, chatId int64) {
	defaultMessage := "Извините, такая команда не поддерживается."
	sendMessage(bot, chatId, defaultMessage)
}

func sendMessage(bot *tgbotapi.BotAPI, chatId int64, message string) {
	msg := tgbotapi.NewMessage(chatId, message)
	bot.Send(msg)
}

func sendPhoto(bot *tgbotapi.BotAPI, chatId int64, photo []byte) {
	msg := tgbotapi.NewPhoto(chatId, tgbotapi.FileBytes{
		Name:  "image",
		Bytes: photo,
	})
	bot.Send(msg)
}

func handleText(bot *tgbotapi.BotAPI, chatId int64, message string) {
	books := request(message)
	if len(books.Items[0].VolmeInfo.ImageLinks.ThubnailImageBytes) != 0 {
		sendPhoto(bot, chatId, books.Items[0].VolmeInfo.ImageLinks.ThubnailImageBytes)
	}
	msg := fmt.Sprintf("Книга: %s\nСписок авторов: %s\nРейтинг: %.1f", books.Items[0].VolmeInfo.Titile, books.Items[0].VolmeInfo.Authors, books.Items[0].VolmeInfo.AverageRating)
	sendMessage(bot, chatId, "Результаты поиска: \n"+msg)
}

func main() {
	bot, err := tgbotapi.NewBotAPI("TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Авторизован на аккаунте %s", bot.Self.UserName)

	req := tgbotapi.NewUpdate(0)
	req.Timeout = 60

	updates := bot.GetUpdatesChan(req)
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Text {
				case "/start":
					startCommand(bot, update.Message.Chat.ID)
				case "/help":
					helpCommand(bot, update.Message.Chat.ID)
				default:
					defaultCommand(bot, update.Message.Chat.ID)
				}
			} else {
				handleText(bot, update.Message.Chat.ID, update.Message.Text)
			}
		}
	}
}
