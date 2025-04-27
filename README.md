# Telegram Bot For Summarization

The project is a telegram bot that:
- Automatically collects news from RSS sources.
- Publishes them into a telegram channel.
- Generates short news summaries using GPT-3.5.
- Managed via commands for administration (adding/editing sources)

## Technology stack

- Go 1.22.4
- Telegram Bot Api
- OpenAI API
- Zap Logger
- PostgreSQL
- Rss parsing
- Docker
- CI/CD: GitHub Actions

## Main libraries

- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) - working with the Telegram API
- [Zap](https://github.com/uber-go/zap) - Logging
- [OpenAI API](https://github.com/sashabaranov/go-openai) - working with the OpenAI API
- [RSS](https://github.com/SlyMarbo/rss) - RSS parsing

1. Clone repository
```bash
git clone https://github.com/lostmyescape/news-tg-bot
```
2 Install dependecies
```bash
go mod tidy
```
3. Launch the database using Docker
```bash
docker-compose up -d
```
4. Configure the application
- Update config.local.hcl with your data
5. Run application:
```bash
go run cmd/main.go
```
## Important Notes
- Only users whose Telegram IDs are listed in the admins array will have access to admin commands.
- Make sure your bot is added to the desired channel/group and has sufficient permissions.


