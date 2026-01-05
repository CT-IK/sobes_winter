package app

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func RegisterAdminHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "yasha_forever", bot.MatchTypeCommand, adminGreetingsHandler)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_admin_home", bot.MatchTypeExact, adminHomeButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_check_entries", bot.MatchTypeExact, checkEntriesButtonHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button_set_slots", bot.MatchTypeExact, setSlotsButtonHandler)
}

func adminChooseContent() (string, *models.InlineKeyboardMarkup) {
	return "Выберите действие:", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Смотреть записи", CallbackData: "button_check_entries"},
				{Text: "Настроить слоты", CallbackData: "button_set_slots"},
			},
		},
	}
}

func checkEntriesContent() (string, *models.InlineKeyboardMarkup) {
	return "Текущие записи:\nTODO", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "button_admin_home"}},
		},
	}
}

func setSlotsContent() (string, *models.InlineKeyboardMarkup) {
	return "Выберите направление:\nTODO", &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "button_admin_home"}},
		},
	}
}

func adminGreetingsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Админский функционал",
		ChatID: update.Message.Chat.ID,
	})

	sendMessage(ctx, b, update.Message.Chat.ID, adminChooseContent)
}

func adminHomeButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)
	editMessage(ctx, b, update.CallbackQuery.Message, adminChooseContent)
}

func checkEntriesButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)
	editMessage(ctx, b, update.CallbackQuery.Message, checkEntriesContent)
}

func setSlotsButtonHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	answerCallback(ctx, b, update)
	editMessage(ctx, b, update.CallbackQuery.Message, setSlotsContent)
}

