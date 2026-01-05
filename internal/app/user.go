package app

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func RegisterUserHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, greetingsHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_home", bot.MatchTypeExact, homeButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_my_profile", bot.MatchTypeExact, myProfileButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_register", bot.MatchTypeExact, registerButtonHandler)
}

func chooseContent() (string, *models.InlineKeyboardMarkup) {
	return "Выберитe действие:", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Моя запись", CallbackData: "button_my_profile"},
				{Text: "Записаться", CallbackData: "button_register"},
			},
		},
	}
}

func myProfileContent() (string, *models.InlineKeyboardMarkup) {
	return "Ваш профиль:\nTODO", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "button_home"}},
		},
	}
}

func registerContent() (string, *models.InlineKeyboardMarkup) {
	return "Введите данные:\nTODO", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "button_home"}},
		},
	}
}

func greetingsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Привет! Это бот для записи на собеседования в комитеты Студенческого Совета Финансового университета.",
		ChatID: update.Message.Chat.ID,
	})

	sendMessage(ctx, b, update.Message.Chat.ID, chooseContent)
}

func homeButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)
	editMessage(ctx, b, update.CallbackQuery.Message, chooseContent)
}

func myProfileButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)
	editMessage(ctx, b, update.CallbackQuery.Message, myProfileContent)
}

func registerButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)
	editMessage(ctx, b, update.CallbackQuery.Message, registerContent)
}
