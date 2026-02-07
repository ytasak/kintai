package slackkintai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

const (
	startText = "リマインダー：業務開始スレ"
	endText   = "リマインダー：業務終了スレ"
)

func ReactStart(ctx context.Context, mode string) error {
	var emoji string
	switch mode {
	case "office":
		emoji = "shussha"
	case "remote":
		emoji = "remote-start"
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
	return reactExact(ctx, startText, emoji)
}

func ReactEnd(ctx context.Context) error {
	return reactExact(ctx, endText, "tai-kin")
}

func reactExact(ctx context.Context, exactText string, emoji string) error {
	token := mustEnv("SLACK_TOKEN")
	ch := mustEnv("SLACK_CHANNEL")

	api := slack.New(token)
	channelID, err := resolveChannelID(api, ch)
	if err != nil {
		return err
	}

	ts, err := findTSByExactText(api, channelID, exactText)
	if err != nil {
		return err
	}

	item := slack.ItemRef{Channel: channelID, Timestamp: ts}
	if err := api.AddReactionContext(ctx, emoji, item); err != nil {
		return fmt.Errorf("reactions.add failed: %w", err)
	}
	return nil
}

func findTSByExactText(api *slack.Client, channelID string, exact string) (string, error) {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	oldest := fmt.Sprintf("%d.000000", startOfDay.Unix())
	latest := fmt.Sprintf("%d.000000", now.Unix())

	var hits []slack.Message
	cursor := ""
	for pages := 0; pages < 5; pages++ {
		hist, err := api.GetConversationHistory(&slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Oldest:    oldest,
			Latest:    latest,
			Inclusive: true,
			Limit:     200,
			Cursor:    cursor,
		})
		if err != nil {
			return "", fmt.Errorf("conversations.history failed: %w", err)
		}
		for _, m := range hist.Messages {
			if m.Text == exact {
				hits = append(hits, m)
			}
		}
		if !hist.HasMore || hist.ResponseMetaData.NextCursor == "" {
			break
		}
		cursor = hist.ResponseMetaData.NextCursor
	}

	if len(hits) == 0 {
		return "", fmt.Errorf("message not found: %q", exact)
	}
	if len(hits) > 1 {
		return "", fmt.Errorf("message not unique: %q hits=%d", exact, len(hits))
	}
	return hits[0].Timestamp, nil
}

func resolveChannelID(api *slack.Client, input string) (string, error) {
	if strings.HasPrefix(input, "C") || strings.HasPrefix(input, "G") {
		return input, nil
	}
	cursor := ""
	for {
		chans, next, err := api.GetConversations(&slack.GetConversationsParameters{
			ExcludeArchived: true,
			Limit:           500,
			Cursor:          cursor,
			Types:           []string{"public_channel", "private_channel"},
		})
		if err != nil {
			return "", fmt.Errorf("conversations.list failed: %w", err)
		}
		for _, c := range chans {
			if c.Name == input {
				return c.ID, nil
			}
		}
		if next == "" {
			break
		}
		cursor = next
	}
	return "", fmt.Errorf("channel not found: %s (set SLACK_CHANNEL to channel ID like Cxxxx)", input)
}

func mustEnv(k string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		panic(errors.New("missing env: " + k))
	}
	return v
}
