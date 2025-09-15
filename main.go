package main

import (
	"fmt"
	"google-book-search-bot/clients/googleBook"
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

func startCommand(bot *tgbotapi.BotAPI, chatID int64) error {
	startMessage := "Введите текст по которому необходимо сделать поиск в Google Books:"
	return sendMessage(bot, chatID, startMessage)
}

func helpCommand(bot *tgbotapi.BotAPI, chatID int64) error {
	helpMessage := "Этот бот делает запрос по введённому тексту в Goole книги и возвращает полученный результат. Присто напишите текст по которому хотите сделать поиск."
	return sendMessage(bot, chatID, helpMessage)
}

func defaultCommand(bot *tgbotapi.BotAPI, chatID int64) error {
	defaultMessage := "Извините, такая команда не поддерживается."
	return sendMessage(bot, chatID, defaultMessage)
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

func newPhotoCaption(books BookView) string {
	index := books.ViewIndex
	book := books.Books.Items[index]
	return fmt.Sprintf(
		"Результаты поиска:\nКнига: %s\nСписок авторов: %s\nРейтинг: %.1f\n[%d/%d]",
		book.VolumeInfo.Title,
		book.VolumeInfo.Authors,
		book.VolumeInfo.AverageRating,
		index+1,
		len(books.Books.Items),
	)
}

func handleText(bot *tgbotapi.BotAPI, chatID int64, message string, booksMap map[int64]BookView, client googleBook.Client) error {
	booksResp, err := client.Books(message)
	if err != nil {
		return fmt.Errorf("books getting error: %w", err)
	}
	if len(booksResp.Items) != 0 {
		var books BookView
		books.Books = booksResp
		books.ViewIndex = 0
		booksMap[chatID] = books

		photo, err := client.BookImage(books.Books.Items[books.ViewIndex].VolumeInfo.ImageLinks)
		if err != nil {
			return fmt.Errorf("photo getting error: %w", err)
		}
		photoConfig := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
			Name:  "image",
			Bytes: photo,
		})
		photoConfig.Caption = newPhotoCaption(books)
		photoConfig.ReplyMarkup = managementKeyboard
		_, err = bot.Send(photoConfig)
		if err != nil {
			return fmt.Errorf("send message error: %w", err)
		}
	} else {
		_, err = bot.Send(tgbotapi.NewMessage(chatID, "По вашему запросу ничего не найдено"))
		if err != nil {
			return fmt.Errorf("send message error: %w", err)
		}
	}

	return nil
}

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, booksMap map[int64]BookView, client googleBook.Client) error {
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID
	data := callback.Data

	books, ok := booksMap[chatID]
	if !ok {
		return fmt.Errorf("cant find books")
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

	photo, err := client.BookImage(books.Books.Items[books.ViewIndex].VolumeInfo.ImageLinks)
	if err != nil {
		return fmt.Errorf("photo getting error: %w", err)
	}

	newPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FileBytes{Name: "imageg", Bytes: photo})
	newPhoto.Caption = newPhotoCaption(books)
	edit := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    chatID,
			MessageID: messageID,
		},
		Media: newPhoto,
	}
	edit.ReplyMarkup = &managementKeyboard

	_, err = bot.Send(edit)
	if err != nil {
		return fmt.Errorf("photo sendong error: %w", err)
	}

	return nil
}

type BookView struct {
	Books     googleBook.GoogleBookResponse
	ViewIndex int
}

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Не передан токен для подклчения к боту. Передайте его первым аргументом командной строки.")
	}
	bot, err := tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Авторизован на аккаунте %s", bot.Self.UserName)

	req := tgbotapi.NewUpdate(0)
	req.Timeout = 60

	updates := bot.GetUpdatesChan(req)
	googleBooksClient := googleBook.New("https://www.googleapis.com/books")
	books := make(map[int64]BookView)
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Text {
				case "/start":
					if err := startCommand(bot, update.Message.Chat.ID); err != nil {
						log.Printf("start command error: %v", err)
					}
				case "/help":
					if err := helpCommand(bot, update.Message.Chat.ID); err != nil {
						log.Printf("help command error: %v", err)
					}
				default:
					if err := defaultCommand(bot, update.Message.Chat.ID); err != nil {
						log.Printf("unknown command error: %v", err)
					}
				}
			} else {
				if err := handleText(bot, update.Message.Chat.ID, update.Message.Text, books, googleBooksClient); err != nil {
					log.Printf("handle text error: %v", err)
				}
			}
		} else if update.CallbackQuery != nil {
			if err := handleCallback(bot, update.CallbackQuery, books, googleBooksClient); err != nil {
				log.Printf("handle callback error: %v", err)
			}
		}

	}
}
