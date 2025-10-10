package prompts

import (
	"fmt"
	"strings"

	"github.com/kriku/kpukbot/internal/models"
)

// ThreadClassificationPrompt generates a prompt for classifying a message into threads
func ThreadClassificationPrompt(message *models.Message, existingThreads []*models.Thread) string {
	var sb strings.Builder

	sb.WriteString("Вы — ассистент по классификации тредов. Проанализируйте следующее сообщение и определите, принадлежит ли оно к какому-либо из существующих тредов обсуждения.\n\n")

	sb.WriteString("Новое сообщение:\n")
	sb.WriteString(fmt.Sprintf("От: %s %s (@%s)\n", message.FirstName, message.LastName, message.Username))
	sb.WriteString(fmt.Sprintf("Текст: %s\n\n", message.Text))

	if len(existingThreads) > 0 {
		sb.WriteString("Существующие треды:\n")
		for i, thread := range existingThreads {
			sb.WriteString(fmt.Sprintf("\nТред %d (ID: %s):\n", i+1, thread.ID))
			sb.WriteString(fmt.Sprintf("Тема: %s\n", thread.Theme))
			sb.WriteString(fmt.Sprintf("Сводка: %s\n", thread.Summary))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Для каждого треда укажите:\n")
	sb.WriteString("1. Вероятность совпадения (от 0.0 до 1.0)\n")
	sb.WriteString("2. Краткое обоснование\n\n")
	sb.WriteString("Если ни один тред не подходит (вероятность для всех ниже 0.5), предложите новую тему для треда.\n\n")

	return sb.String()
}

// ThreadSummaryPrompt generates a prompt for summarizing a thread
func ThreadSummaryPrompt(messages []*models.Message) string {
	var sb strings.Builder

	sb.WriteString("Создайте краткую сводку для следующего треда обсуждения. Определите основную тему и предоставьте краткое изложение.\n\n")
	sb.WriteString("Сообщения:\n")

	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("\n%d. %s %s: %s\n", i+1, msg.FirstName, msg.LastName, msg.Text))
	}

	sb.WriteString("\nУкажите:\n")
	sb.WriteString("1. Краткую тему (5-10 слов)\n")
	sb.WriteString("2. Краткую сводку (2-3 предложения)\n\n")

	return sb.String()
}

// ResponseAnalysisPrompt generates a prompt for analyzing if a response is needed
func ResponseAnalysisPrompt(thread *models.Thread, messages []*models.Message, newMessage *models.Message) string {
	var sb strings.Builder

	sb.WriteString("Вы — ассистент-бот, анализирующий, требуется ли ответ в этом обсуждении.\n\n")

	sb.WriteString(fmt.Sprintf("Тема треда: %s\n", thread.Theme))
	sb.WriteString(fmt.Sprintf("Сводка треда: %s\n\n", thread.Summary))

	sb.WriteString("Последние сообщения:\n")
	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, msg.FirstName, msg.Text))
	}

	sb.WriteString(fmt.Sprintf("\nНовое сообщение: %s: %s\n\n", newMessage.FirstName, newMessage.Text))

	sb.WriteString("Проанализируйте, должен ли бот отвечать. Учитывайте:\n")
	sb.WriteString("1. Есть ли вопрос, требующий ответа?\n")
	sb.WriteString("2. Требуется ли проверка фактов?\n")

	return sb.String()
}

func FactCheckingNeedsPrompt(context string, statement string) string {
	var sb strings.Builder

	sb.WriteString("Вы — ассистент по проверке фактов. Проанализируйте, требует ли следующее сообщение проверки фактов.\n\n")
	sb.WriteString(fmt.Sprintf("Контекст: %s\n\n", context))

	sb.WriteString(fmt.Sprintf("Утверждение: %s\n\n", statement))

	sb.WriteString("Проанализируйте, требует ли это сообщение проверки фактов. Учитывайте:\n")
	sb.WriteString("1. Содержит ли сообщение утверждения, которые можно проверить?\n")
	sb.WriteString("2. Являются ли эти утверждения спорными или потенциально ложными?\n")

	return sb.String()
}

// FactCheckingPrompt generates a prompt for fact-checking
func FactCheckingPrompt(context string, statement string) string {
	return fmt.Sprintf(`Вы — ассистент по проверке фактов. Проанализируйте следующее утверждение в контексте.

Контекст: %s

Утверждение для проверки: %s

Предоставьте:
1. Результат проверки (истина/ложь/неопределенно)
2. Объяснение
3. Дополнительный контекст, если необходимо
`, context, statement)
}

// GeneralResponsePrompt generates a prompt for general responses
func GeneralResponsePrompt(thread *models.Thread, messages []*models.Message, newMessage *models.Message) string {
	var sb strings.Builder

	sb.WriteString("Сгенерируйте полезный и контекстуально уместный ответ.\n\n")
	sb.WriteString(fmt.Sprintf("Тред: %s\n\n", thread.Theme))

	sb.WriteString("Обсуждение:\n")
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.FirstName, msg.Text))
	}
	sb.WriteString(fmt.Sprintf("\n%s: %s\n\n", newMessage.FirstName, newMessage.Text))

	sb.WriteString("Предоставьте краткий, полезный ответ. Будьте дружелюбны, но профессиональны.")

	return sb.String()
}
