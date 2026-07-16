# Система отслеживания статусов пользователей

## Архитектура

Система использует Redis для хранения статусов пользователей с TTL (Time To Live) для автоматической очистки устаревших записей.

### Структура данных в Redis:

1. **user:status:{userID}** - JSON с текущим статусом пользователя
   - Значения: `online`, `offline`, `typing:{chatID}`
   - TTL: 30 секунд для online, 10 секунд для typing

2. **user:connections:{userID}** - Set с connection IDs
   - Хранит все активные WebSocket подключения пользователя
   - Пользователь считается онлайн, пока есть хотя бы одно подключение

3. **user:active_chat:{userID}** - ID активного чата
   - TTL: 30 секунд

4. **chat:typing:{chatID}** - Hash с пользователями, которые печатают
   - Ключ: userID
   - Значение: JSON с timestamp и chatID
   - TTL: 10 секунд на уровне приложения

## WebSocket операции

### Установка статуса "печатает"
```json
{
  "operation": "typing_start",
  "message": {
    "sender_id": "user-uuid",
    "chat_id": "chat-uuid"
  }
}
```

### Остановка печатания
```json
{
  "operation": "typing_stop", 
  "message": {
    "sender_id": "user-uuid",
    "chat_id": "chat-uuid"
  }
}
```

### Получение статусов печатающих в чате
```json
{
  "operation": "get_status",
  "chat_id": "chat-uuid"
}
```

### Уведомление о статусе печатания (рассылается всем участникам чата)
```json
{
  "operation": "typing_update",
  "status": "typing",  // или "stopped_typing"
  "chat_id": "chat-uuid",
  "message": {
    "sender_id": "user-uuid",
    "chat_id": "chat-uuid"
  }
}
```

## HTTP API

### Получение статуса пользователя
```
GET /api/v1/status/user/{user_id}
```

**Ответ:**
```json
{
  "user_id": "uuid",
  "status": "online|offline|typing:{chatID}",
  "chat_id": "uuid",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Получение статусов нескольких пользователей
```
POST /api/v1/status/users
```

**Запрос:**
```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**Ответ:**
```json
{
  "statuses": {
    "uuid1": {
      "user_id": "uuid1",
      "status": "online",
      "timestamp": "2024-01-01T00:00:00Z"
    },
    "uuid2": {
      "user_id": "uuid2", 
      "status": "typing:chat-uuid",
      "chat_id": "chat-uuid",
      "timestamp": "2024-01-01T00:00:00Z"
    }
  }
}
```

### Получение списка печатающих в чате
```
GET /api/v1/status/chat/{chat_id}/typing
```

**Ответ:**
```json
[
  {
    "user_id": "uuid",
    "status": "typing:chat-uuid",
    "chat_id": "chat-uuid",
    "timestamp": "2024-01-01T00:00:00Z"
  }
]
```

### Получение активного чата пользователя
```
GET /api/v1/status/user/{user_id}/active-chat
```

**Ответ:**
```json
{
  "user_id": "uuid",
  "chat_id": "chat-uuid"
}
```

## Логика работы

### Подключение пользователя
1. При установке WebSocket соединения вызывается `SetUserOnline`
2. Session ID добавляется в список подключений пользователя
3. Статус устанавливается в `online` с TTL 30 секунд

### Отключение пользователя
1. При закрытии WebSocket соединения вызывается `SetUserOffline`
2. Session ID удаляется из списка подключений
3. Если подключений не осталось, статус меняется на `offline`

### Печатание в чате
1. При начале печатания отправляется `typing_start`
2. Статус меняется на `typing:{chatID}` с TTL 10 секунд
3. Все участники чата получают уведомление `typing_update`
4. При остановке печатания отправляется `typing_stop`
5. Статус возвращается к `online`

### Переключение между чатами
1. При смене активного чата отправляется `switch_community`
2. Активный чат сохраняется в Redis с TTL 30 секунд

## Настройка таймаутов

- **Online TTL**: 30 секунд - если пользователь не обновляет статус, он считается оффлайн
- **Typing TTL**: 10 секунд - если пользователь перестал печатать, статус автоматически сбрасывается
- **Active Chat TTL**: 30 секунд - если пользователь неактивен, чат забывается

## Масштабирование

Система поддерживает горизонтальное масштабирование:
- Redis как центральное хранилище статусов
- Множество экземпляров приложения могут работать с одним Redis
- WebSocket соединения распределяются между инстансами

## Мониторинг

Рекомендуется мониторить:
- Количество онлайн пользователей
- Количество активных WebSocket соединений
- Использование памяти Redis
- Latency операций с Redis