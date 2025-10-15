package prompts

import (
	"fmt"
	"strings"

	"github.com/kriku/kpukbot/internal/constants"
	"github.com/kriku/kpukbot/internal/models"
)

// ThreadClassificationPrompt generates a prompt for classifying a message into threads
func ThreadClassificationPrompt(message *models.Message, existingThreads []*models.Thread) string {
	var sb strings.Builder

	sb.WriteString("You are a thread classification assistant. Analyze the following message and determine if it belongs to any of the existing discussion threads.\n\n")

	sb.WriteString("New message:\n")
	sb.WriteString(fmt.Sprintf("From: %s %s (@%s)\n", message.FirstName, message.LastName, message.Username))
	sb.WriteString(fmt.Sprintf("Text: %s\n\n", message.Text))

	if len(existingThreads) > 0 {
		sb.WriteString("Existing threads:\n")
		for i, thread := range existingThreads {
			sb.WriteString(fmt.Sprintf("\nThread %d (ID: %s):\n", i+1, thread.ID))
			sb.WriteString(fmt.Sprintf("Topic: %s\n", thread.Theme))
			sb.WriteString(fmt.Sprintf("Summary: %s\n", thread.Summary))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("For each thread, specify:\n")
	sb.WriteString("1. Match probability (from 0.0 to 1.0)\n")
	sb.WriteString("2. Brief justification\n\n")
	sb.WriteString(
		fmt.Sprintf(
			"If no thread matches (probability for all is below 0.5), suggest a new thread topic, no longer than %d characters.\n\n",
			constants.MaxThreadThemeLength,
		),
	)

	return sb.String()
}

// ThreadSummaryPrompt generates a prompt for summarizing a thread
func ThreadSummaryPrompt(messages []*models.Message) string {
	var sb strings.Builder

	sb.WriteString("Create a brief summary for the following discussion thread. Identify the main topic and provide a concise overview.\n\n")
	sb.WriteString("Messages:\n")

	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("\n%d. %s %s: %s\n", i+1, msg.FirstName, msg.LastName, msg.Text))
	}

	sb.WriteString("\nSpecify:\n")
	sb.WriteString("1. Brief topic (5-10 words)\n")
	sb.WriteString("2. Concise summary (2-3 sentences)\n\n")

	return sb.String()
}

// ResponseAnalysisPrompt generates a prompt for analyzing if a response is needed
func ResponseAnalysisPrompt(thread *models.Thread, messages []*models.Message, newMessage *models.Message) string {
	var sb strings.Builder

	sb.WriteString("You are an assistant bot analyzing whether a response is needed in this discussion.\n\n")

	sb.WriteString(fmt.Sprintf("Thread topic: %s\n", thread.Theme))
	sb.WriteString(fmt.Sprintf("Thread summary: %s\n\n", thread.Summary))

	sb.WriteString("Recent messages:\n")
	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, msg.FirstName, msg.Text))
	}

	sb.WriteString(fmt.Sprintf("\nNew message: %s: %s\n\n", newMessage.FirstName, newMessage.Text))

	sb.WriteString("Analyze whether the bot should respond. Consider:\n")
	sb.WriteString("1. Is there a question that requires an answer?\n")
	sb.WriteString("2. Is fact-checking needed?\n")

	return sb.String()
}

// GeneralResponsePrompt generates a prompt for general responses
func GeneralResponsePrompt(thread *models.Thread, messages []*models.Message, newMessage *models.Message) string {
	var sb strings.Builder

	sb.WriteString("Generate a helpful and contextually appropriate response.\n\n")
	sb.WriteString(fmt.Sprintf("Thread: %s\n\n", thread.Theme))

	sb.WriteString("Discussion:\n")
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.FirstName, msg.Text))
	}
	sb.WriteString(fmt.Sprintf("\n%s: %s\n\n", newMessage.FirstName, newMessage.Text))

	sb.WriteString("Provide a brief, helpful response. Be friendly but professional.")

	return sb.String()
}
