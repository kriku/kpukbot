Telegram Bot + AI = ❤️
-----------------------------

This is a simple Telegram Bot that uses Google Gemini model to generate responses to user messages. The bot is built using the Go programming language and the Functions Framework for Go.

Google Cloud Run Function is called by a Telegram Webhook with updates from the Telegram Bot API. The bot uses the Google Gemini model to generate responses to user messages.

## Getting Started


### Local run

Run Locally in polling mode with Go installed:
```
go run cmd/main.go
```

Run Locally with pack & Docker:
```
pack build --builder=gcr.io/buildpacks/builder sample-functions-framework-go
docker run -p8080:8080 sample-functions-framework-go
```

### Production

Google Cloud Run integrated with this repository.

After build cloud function is called by a Telegram Webhook with updates from the Telegram Bot API. The bot uses the Google Gemini model to generate responses to user messages.
