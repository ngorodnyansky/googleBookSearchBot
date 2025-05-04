package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var managementKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("<<", "previous"),
		tgbotapi.NewInlineKeyboardButtonData(">>", "next"),
	),
)

func startCommand(bot *tgbotapi.BotAPI, chatID int64) {
	startMessage := "Введите текст по которому необходимо сделать поиск в Google Books:"
	sendMessage(bot, chatID, startMessage)
}

func helpCommand(bot *tgbotapi.BotAPI, chatID int64) {
	helpMessage := "Этот бот делает запрос по введённому тексту в Goole книги и возвращает полученный результат. Присто напишите текст по которому хотите сделать поиск."
	sendMessage(bot, chatID, helpMessage)
}

func defaultCommand(bot *tgbotapi.BotAPI, chatID int64) {
	defaultMessage := "Извините, такая команда не поддерживается."
	sendMessage(bot, chatID, defaultMessage)
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}

func newPhotoCaption(books BookView) string {
	index := books.ViewIndex
	book := books.Books.Items[index]
	return fmt.Sprintf(
		"Результаты поиска:\nКнига: %s\nСписок авторов: %s\nРейтинг: %.1f\n[%d/%d]",
		book.VolmeInfo.Titile,
		book.VolmeInfo.Authors,
		book.VolmeInfo.AverageRating,
		index+1,
		len(books.Books.Items),
	)
}

func getImage(imageInBytes []byte) []byte {
	if len(imageInBytes) == 0 {
		image, _ := os.ReadFile("image_not_found.png")
		return image
	}
	return imageInBytes
}

func handleText(bot *tgbotapi.BotAPI, chatID int64, message string, booksMap map[int64]BookView) {
	booksResp := request(message)
	if len(booksResp.Items) != 0 {
		var books BookView
		books.Books = booksResp
		books.ViewIndex = 0
		booksMap[chatID] = books

		photoConfig := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
			Name:  "image",
			Bytes: getImage(books.Books.Items[books.ViewIndex].VolmeInfo.ImageLinks.ThubnailImageBytes),
		})
		photoConfig.Caption = newPhotoCaption(books)
		photoConfig.ReplyMarkup = managementKeyboard
		bot.Send(photoConfig)
	} else {
		bot.Send(tgbotapi.NewMessage(chatID, "По вашему запросу ничего не найдено"))
	}
}

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, booksMap map[int64]BookView) {
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID
	data := callback.Data

	books, ok := booksMap[chatID]
	if !ok {
		return
	}

	switch data {
	case "next":
		if books.ViewIndex+1 == len(books.Books.Items) {
			books.ViewIndex = 0
		} else {
			books.ViewIndex += 1
		}
	case "previous":
		if books.ViewIndex == 0 {
			books.ViewIndex = len(books.Books.Items) - 1
		} else {
			books.ViewIndex -= 1
		}
	}
	booksMap[chatID] = books

	deleteConfig := tgbotapi.NewDeleteMessage(chatID, messageID)
	bot.Send(deleteConfig)

	photoConfig := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
		Name:  "image",
		Bytes: getImage(books.Books.Items[books.ViewIndex].VolmeInfo.ImageLinks.ThubnailImageBytes),
	})
	photoConfig.Caption = newPhotoCaption(books)
	photoConfig.ReplyMarkup = managementKeyboard
	bot.Send(photoConfig)
}

type BookView struct {
	Books     GoogleBookResponce
	ViewIndex int
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
	books := make(map[int64]BookView)
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
				handleText(bot, update.Message.Chat.ID, update.Message.Text, books)
			}
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery, books)
		}

	}
}
