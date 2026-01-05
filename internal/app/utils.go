package app

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func answerCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
}

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, contentFunc func() (string, *models.InlineKeyboardMarkup)) {
	text, kb := contentFunc()

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func editMessage(ctx context.Context, b *bot.Bot, message models.MaybeInaccessibleMessage, contentFunc func() (string, *models.InlineKeyboardMarkup)) {
	if message.Type == models.MaybeInaccessibleMessageTypeInaccessibleMessage {
		return
	}

	text, kb := contentFunc()
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      message.Message.Chat.ID,
		MessageID:   message.Message.ID,
		Text:        text,
		ReplyMarkup: kb,
	})
}
