# ПРОСТАЯ архитектура статусов с Redis

## 🎯 Что мы сделали:

### 1. **Добавили Redis клиент прямо в Hub:**
```go
type Hub struct {
    clients map[string]*Client      // WebSocket клиенты в памяти
    redis   *redis.Client           // Redis для статусов
    // ... остальное
}
```

### 2. **Убрали все сложности:**
- ❌ Нет отдельного `StatusService`
- ❌ Нет HTTP API для статусов  
- ❌ Нет сложных структур
- ✅ Только Redis команды в Hub

## 🔧 Как это работает:

### 1. **При подключении (handshake):**
```go
func (h *Hub) onConnect(ctx context.Context, client *Client, subscribeUsers []uuid.UUID) {
    // 1. Сохраняем клиента
    h.clients[client.ID] = client
    
    // 2. Устанавливаем онлайн в Redis
    h.setUserOnline(ctx, userID)
    
    // 3. Добавляем подписки (за кем следит пользователь)
    for _, subUserID := range subscribeUsers {
        h.addUserSubscription(ctx, userID, subUserID)
    }
    
    // 4. Отправляем текущие статусы подписанных пользователей
    go h.sendInitialStatuses(ctx, userID, client)
}
```

### 2. **Redis хранит:**
```
user:online:{userID}       = timestamp      ← онлайн ли
user:active:{userID}       = chatID         ← в каком чате
user:typing:{userID}       = chatID         ← где печатает
user:subs:{userID}         = [user1, user2] ← за кем следит
```

### 3. **WebSocket операции:**
```javascript
// Клиент сообщает что делает:
ws.send({operation: "switch_community", chat_id: "general"})  // Зашел в чат
ws.send({operation: "typing_start", chat_id: "general"})     // Начал печатать
ws.send({operation: "typing_stop", chat_id: "general"})      // Перестал печатать

// Получает уведомления:
{"operation": "user_status_update", user_id: "vasya", status: "typing_start", chat_id: "general"}
{"operation": "user_status_update", user_id: "vasya", status: "active_chat_changed", chat_id: "random"}
```

## 🎮 Пример потока:

**Вася подключается:**
1. WebSocket handshake → `session-abc`
2. Hub.onConnect() → Redis: `SET user:online:vasya <timestamp>`
3. Добавляет подписки: `SADD user:subs:vasya petya, masha`
4. Отправляет Васе статусы Пети и Маши

**Вася заходит в #general:**
1. Клиент: `{"operation": "switch_community", "chat_id": "general"}`
2. Сервер: `SET user:active:vasya "general"`
3. Уведомляет подписчиков (Петя и Маша)

**Вася печатает в #general:**
1. Клиент: `{"operation": "typing_start", "chat_id": "general"}`
2. Сервер: `SET user:typing:vasya "general" EX 10`
3. Уведомляет подписчиков

## 📁 Созданные файлы:

1. **`internal/services/http/messages.go`** - обновленный Hub с Redis
2. **`internal/services/http/redis_status.go`** - простые Redis методы
3. **`internal/services/http/services.go`** - убрали StatusService
4. **`cmd/messenger/main.go`** - обновленная инициализация

## 🚀 Преимущества:

1. **Просто**: Redis команды прямо в Hub
2. **Минимально**: Нет лишних абстракций
3. **Эффективно**: Все в одном месте
4. **Понятно**: Легко читать и поддерживать

## 🎯 ИТОГ:

Теперь у тебя:
- **Hub** с WebSocket клиентами и Redis
- **Redis** хранит все статусы и подписки
- **WebSocket** для всего общения
- **Просто** и **понятно** 🎉