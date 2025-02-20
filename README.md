# pgtk-schedule

## Конфигурация

### Бот

Конфигурация указывается в файле `.env`

| Название    | Описание                                                 | Тип      | Обязательно | Стандартное значение |
| ----------- | -------------------------------------------------------- | -------- | ----------- | -------------------- |
| `BOT_TOKEN` | Токен от бота в ТГ                                       | `string` | [x]         | -                    |
| `ADMIN_ID`  | ID пользователя, который будет иметь роль администратора | `int64`  | [x]         | -                    |
| `DB_CONN`   | Строка для подключения к PostgreSQL                      | `string` | [x]         | -                    |

### Docker

Для докера в файле `.env` дополнительно необходимо добавить параметры `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_PORT`.

Пример запуска:
```sh
docker compose --env-file=.env up -d
```

## Миграции

> Необходимо заменить `host`, `user`, `password`, `port`, `database`. Дополнительно можно указать `search_path`.

### Применить миграции

```sh
go tool goose postgres "host=localhost user=u password=pwd port=5432 database=pgtk" up -dir migrations
```

### Откатить миграции

```sh
go tool goose postgres "host=localhost user=u password=pwd port=5432 database=pgtk" down -dir migrations
```

## Команды администратора

### /send

Позволяет отправить всем пользователям сообщение.

Пример использования:
```
/send Привет, мир!
```