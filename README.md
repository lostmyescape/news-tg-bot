# Телеграм-бот для автоматической сборки новостей

Проект построен на базе Go, который:
- Автоматически собирает новости из RSS-источников.
- Публикует их в телеграм-канале.
- Генерирует саммари новостей с помощью GPT-3.5.
- Управляется с помощью команд для администрирования (добавление/редактирование источников)

## Требования

- Go 1.23.4
- PostgreSQL
- Docker

## Основные библиотеки

- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) - работа с Telegram API
- [Zap](https://github.com/uber-go/zap) - Логгирование
- [OpenAI API](https://github.com/sashabaranov/go-openai) - работа с OpenAI API
- [RSS](https://github.com/SlyMarbo/rss) - RSS парсинг

## Запуск проекта

1. Скопируйте репозиторий
```bash
git clone https://github.com/lostmyescape/news-tg-bot
```
2 Обновите зависимости
```bash
go mod tidy
```
3. Запустите базу данных в Docker
```bash
docker-compose up -d
```
4. Настройка приложения
- Обновите config.local.hcl: добавьте свой telegram id в поле "Администраторы"
5. Запустите приложение:
```bash
go run cmd/main.go
```
6. Ссылки на бота и канал
- [telegram bot](https://t.me/nnnewsfeed_bot)
- [telegram channel](https://t.me/golangnewsbott)

## Важно!
- Только пользователи, чьи идентификаторы Telegram указаны в списке администраторов, будут иметь доступ к командам администратора.
- Убедитесь, что ваш бот добавлен в нужный канал/группу и имеет достаточные права доступа.


