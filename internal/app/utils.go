package app

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func answerCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
}

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string, kb *models.InlineKeyboardMarkup) {
	if kb == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   text,
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func editMessage(ctx context.Context, b *bot.Bot, message models.MaybeInaccessibleMessage, text string, kb *models.InlineKeyboardMarkup) {
	if message.Type == models.MaybeInaccessibleMessageTypeInaccessibleMessage {
		return
	}

	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      message.Message.Chat.ID,
		MessageID:   message.Message.ID,
		Text:        text,
		ReplyMarkup: kb,
	})

	if err != nil {
		log.Fatal(err)
	}
}
