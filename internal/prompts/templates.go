package prompts

import (
	"encoding/json"
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
	sb.WriteString("2. Is user introduced himself?\n")

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

// IntroductionAnalysisPrompt generates a prompt for analyzing if a message is an introduction
func IntroductionAnalysisPrompt(message *models.Message) string {
	var sb strings.Builder

	sb.WriteString("Analyze whether the following message is a user introduction. Return a JSON object with the following structure:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"is_introduction\": true/false,\n")
	sb.WriteString("  \"confidence\": 0.0-1.0,\n")
	sb.WriteString("  \"reasoning\": \"Brief explanation of your analysis\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("Guidelines for detecting introductions:\n")
	sb.WriteString("- User shares personal information about themselves\n")
	sb.WriteString("- User mentions their name, profession, interests, or hobbies\n")
	sb.WriteString("- User uses phrases like 'I am', 'My name is', 'I like', 'I enjoy', etc.\n")
	sb.WriteString("- User expresses intent to introduce themselves or share about themselves\n")
	sb.WriteString("- User greets while sharing personal details\n\n")

	sb.WriteString("NOT introductions:\n")
	sb.WriteString("- Simple greetings without personal information\n")
	sb.WriteString("- Questions or casual conversation\n")
	sb.WriteString("- Comments about external topics\n")
	sb.WriteString("- Responses to others without self-disclosure\n\n")

	sb.WriteString(fmt.Sprintf("Message from %s:\n", message.FirstName))
	sb.WriteString(fmt.Sprintf("\"%s\"\n", message.Text))

	return sb.String()
}

// UserInformationExtractionPrompt generates a prompt for extracting user information from introduction messages
func UserInformationExtractionPrompt(message *models.Message) string {
	var sb strings.Builder

	sb.WriteString("Extract user information from the following introduction message. Return a JSON object with the following structure:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"bio\": \"Brief bio or self-description if mentioned\",\n")
	sb.WriteString("  \"interests\": [\"interest1\", \"interest2\"],\n")
	sb.WriteString("  \"hobbies\": [\"hobby1\", \"hobby2\"]\n")
	sb.WriteString("}\n\n")

	sb.WriteString("Rules:\n")
	sb.WriteString("- Only extract explicitly mentioned information\n")
	sb.WriteString("- Bio should be 1-2 sentences maximum\n")
	sb.WriteString("- Interests are topics they're curious about or study\n")
	sb.WriteString("- Hobbies are activities they do in their free time\n")
	sb.WriteString("- Use empty string for bio and empty arrays for interests/hobbies if not mentioned\n\n")

	sb.WriteString(fmt.Sprintf("Message from %s %s:\n", message.FirstName, message.LastName))
	sb.WriteString(fmt.Sprintf("\"%s\"\n", message.Text))

	return sb.String()
}

// IntroductionConfirmationPrompt generates a prompt for creating confirmation responses
func IntroductionConfirmationPrompt(message *models.Message, userInfo *models.UserInformation) string {
	var sb strings.Builder

	sb.WriteString("Generate a warm, friendly confirmation message for a user's introduction. ")
	sb.WriteString("Acknowledge what they shared and welcome them to the community.\n\n")

	sb.WriteString(fmt.Sprintf("User: %s %s\n", message.FirstName, message.LastName))
	sb.WriteString(fmt.Sprintf("Original message: \"%s\"\n\n", message.Text))

	if userInfo.Bio != "" {
		sb.WriteString(fmt.Sprintf("Bio: %s\n", userInfo.Bio))
	}

	if len(userInfo.Interests) > 0 {
		sb.WriteString(fmt.Sprintf("Interests: %s\n", strings.Join(userInfo.Interests, ", ")))
	}

	if len(userInfo.Hobbies) > 0 {
		sb.WriteString(fmt.Sprintf("Hobbies: %s\n", strings.Join(userInfo.Hobbies, ", ")))
	}

	sb.WriteString("\nGuidelines:\n")
	sb.WriteString("- Be warm and welcoming\n")
	sb.WriteString("- Acknowledge specific interests/hobbies they mentioned\n")
	sb.WriteString("- Keep it personal but not too lengthy\n")
	sb.WriteString("- Use their first name\n")
	sb.WriteString("- Maximum 300 characters\n")

	return sb.String()
}

// ParseUserInformationResponse parses the JSON response from user information extraction
func ParseUserInformationResponse(jsonResponse string) (*models.UserInformation, error) {
	var userInfo models.UserInformation
	err := json.Unmarshal([]byte(jsonResponse), &userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user information JSON: %w", err)
	}
	return &userInfo, nil
}

// ParseIntroductionAnalysisResponse parses the JSON response from introduction analysis
func ParseIntroductionAnalysisResponse(jsonResponse string) (*models.IntroductionAnalysisResult, error) {
	var result models.IntroductionAnalysisResult
	err := json.Unmarshal([]byte(jsonResponse), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse introduction analysis JSON: %w", err)
	}
	return &result, nil
}
