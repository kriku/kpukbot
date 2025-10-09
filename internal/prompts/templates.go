package prompts

import (
	"fmt"
	"strings"

	"github.com/kriku/kpukbot/internal/models"
)

// ThreadClassificationPrompt generates a prompt for classifying a message into threads
func ThreadClassificationPrompt(message *models.Message, existingThreads []*models.Thread) string {
	var sb strings.Builder

	sb.WriteString("You are a thread classification assistant. Analyze the following message and determine if it belongs to any existing conversation threads.\n\n")

	sb.WriteString(fmt.Sprintf("New Message:\n"))
	sb.WriteString(fmt.Sprintf("From: %s %s (@%s)\n", message.FirstName, message.LastName, message.Username))
	sb.WriteString(fmt.Sprintf("Text: %s\n\n", message.Text))

	if len(existingThreads) > 0 {
		sb.WriteString("Existing Threads:\n")
		for i, thread := range existingThreads {
			sb.WriteString(fmt.Sprintf("\nThread %d (ID: %s):\n", i+1, thread.ID))
			sb.WriteString(fmt.Sprintf("Theme: %s\n", thread.Theme))
			sb.WriteString(fmt.Sprintf("Summary: %s\n", thread.Summary))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("For each thread, provide:\n")
	sb.WriteString("1. Match probability (0.0 to 1.0)\n")
	sb.WriteString("2. Brief reasoning\n\n")
	sb.WriteString("If no threads match well (all below 0.5), suggest a new thread theme.\n\n")

	return sb.String()
}

// ThreadSummaryPrompt generates a prompt for summarizing a thread
func ThreadSummaryPrompt(messages []*models.Message) string {
	var sb strings.Builder

	sb.WriteString("Summarize the following conversation thread. Identify the main theme and provide a brief summary.\n\n")
	sb.WriteString("Messages:\n")

	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("\n%d. %s %s: %s\n", i+1, msg.FirstName, msg.LastName, msg.Text))
	}

	sb.WriteString("\nProvide:\n")
	sb.WriteString("1. A concise theme (5-10 words)\n")
	sb.WriteString("2. A brief summary (2-3 sentences)\n\n")

	return sb.String()
}

// ResponseAnalysisPrompt generates a prompt for analyzing if a response is needed
func ResponseAnalysisPrompt(thread *models.Thread, messages []*models.Message, newMessage *models.Message) string {
	var sb strings.Builder

	sb.WriteString("You are a bot assistant analyzing whether a response is needed for this conversation.\n\n")

	sb.WriteString(fmt.Sprintf("Thread Theme: %s\n", thread.Theme))
	sb.WriteString(fmt.Sprintf("Thread Summary: %s\n\n", thread.Summary))

	sb.WriteString("Recent Messages:\n")
	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, msg.FirstName, msg.Text))
	}

	sb.WriteString(fmt.Sprintf("\nNew Message: %s: %s\n\n", newMessage.FirstName, newMessage.Text))

	sb.WriteString("Analyze if the bot should respond. Consider:\n")
	sb.WriteString("1. Is there a question that needs answering?\n")
	sb.WriteString("2. Is fact-checking needed?\n")
	sb.WriteString("3. Should agreements or decisions be tracked?\n")
	sb.WriteString("4. Is a reminder appropriate?\n")
	sb.WriteString("5. Is clarification needed?\n\n")

	return sb.String()
}

// FactCheckingPrompt generates a prompt for fact-checking
func FactCheckingPrompt(context string, statement string) string {
	return fmt.Sprintf(`You are a fact-checking assistant. Analyze the following statement in context.

Context: %s

Statement to verify: %s

Provide:
1. Verification result (true/false/uncertain)
2. Explanation
3. Additional context if needed
`, context, statement)
}

// ReminderPrompt generates a prompt for creating reminders
func ReminderPrompt(thread *models.Thread, messages []*models.Message) string {
	var sb strings.Builder

	sb.WriteString("Extract any commitments, deadlines, or action items that should be tracked.\n\n")
	sb.WriteString(fmt.Sprintf("Thread: %s\n\n", thread.Theme))

	sb.WriteString("Messages:\n")
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", msg.FirstName, msg.Text))
	}

	return sb.String()
}

// AgreementTrackingPrompt generates a prompt for tracking agreements
func AgreementTrackingPrompt(thread *models.Thread, messages []*models.Message) string {
	var sb strings.Builder

	sb.WriteString("Identify any agreements, decisions, or consensus reached in this conversation.\n\n")
	sb.WriteString(fmt.Sprintf("Thread: %s\n\n", thread.Theme))

	sb.WriteString("Messages:\n")
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", msg.FirstName, msg.Text))
	}

	return sb.String()
}

// GeneralResponsePrompt generates a prompt for general responses
func GeneralResponsePrompt(thread *models.Thread, messages []*models.Message, newMessage *models.Message) string {
	var sb strings.Builder

	sb.WriteString("Generate a helpful and contextually appropriate response.\n\n")
	sb.WriteString(fmt.Sprintf("Thread: %s\n\n", thread.Theme))

	sb.WriteString("Conversation:\n")
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.FirstName, msg.Text))
	}
	sb.WriteString(fmt.Sprintf("\n%s: %s\n\n", newMessage.FirstName, newMessage.Text))

	sb.WriteString("Provide a concise, helpful response. Be friendly but professional.")

	return sb.String()
}
