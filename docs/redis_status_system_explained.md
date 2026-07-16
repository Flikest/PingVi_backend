# Объяснение системы статусов с Redis - ПРОСТЫМИ СЛОВАМИ

## 🎯 Что мы хотим сделать?

Мы хотим чтобы:
1. **Видеть кто онлайн/оффлайн** - зеленая/серая точка у аватарки
2. **Видеть в каком чате сидит пользователь** - "Вася в чате #general"
3. **Видеть кто печатает в чате** - "Вася печатает..."

## 🏗️ Архитектура - КАК ЭТО РАБОТАЕТ?

### 1. Redis - это БЫСТРАЯ ПАМЯТЬ (как оперативка)

Представь что Redis - это **доска с заметками**:
- Быстро пишешь заметки (SET)
- Быстро читаешь заметки (GET)
- Заметки сами исчезают через время (TTL)

### 2. Какие "заметки" мы храним:

```
КЛЮЧ (адрес заметки)              ЗНАЧЕНИЕ (что пишем)           ЧЕРЕЗ СКОЛЬКО УДАЛИТЬ
──────────────────────────────────────────────────────────────────────────────────────
user:status:123                  {"status":"online"}            30 секунд
user:connections:123             [ws-1, ws-2]                   (без TTL)
user:active_chat:123             "chat-456"                     5 минут  
user:typing_chat:123             "chat-789"                     10 секунд
chat:online_users:456            [user-123, user-789]           (без TTL)
chat:typing_users:789            [user-123]                     (без TTL)
```

### 3. Поток данных - ЧТО ПРОИСХОДИТ:

**Когда Вася открывает приложение:**
```
Браузер Васи → WebSocket → Go сервер → Redis:
1. SET user:status:vasya "online" EX 30
2. SADD user:connections:vasya "ws-session-1"
```

**Когда Вася заходит в чат #general:**
```
Браузер: {"operation": "switch_chat", "chat_id": "general"}
Go сервер → Redis:
1. SET user:active_chat:vasya "general" EX 300
2. SADD chat:online_users:general "vasya"
```

**Когда Вася начинает печатать в #random:**
```
Браузер: {"operation": "typing_start", "chat_id": "random"}
Go сервер → Redis:
1. SET user:typing_chat:vasya "random" EX 10
2. SADD chat:typing_users:random "vasya"
3. Рассылаем всем в #random: "Вася печатает..."
```

**Когда Петя хочет узнать где Вася:**
```
Браузер Пети → HTTP GET /api/v1/status/user/vasya/presence
Go сервер → Redis:
1. GET user:status:vasya → "online" (если не истек 30 сек)
2. GET user:active_chat:vasya → "general"
3. GET user:typing_chat:vasya → "random"
Ответ: {"online": true, "active_chat": "general", "typing_chat": "random"}
```

## 🔧 Ключевые команды Redis (самые важные):

```bash
# Записать значение с временем жизни
SET user:status:123 "online" EX 30

# Прочитать значение
GET user:status:123

# Добавить в список (Set)
SADD user:connections:123 "ws-abc"

# Удалить из списка
SREM user:connections:123 "ws-abc"

# Посмотреть сколько в списке
SCARD user:connections:123

# Посмотреть всех в списке
SMEMBERS chat:online_users:456

# Узнать когда удалится ключ
TTL user:status:123
```

## 🎮 Пример работы в реальном времени:

### Сценарий 1: Вася онлайн в двух вкладках
```
Вкладка 1: SADD user:connections:vasya "ws-tab1"
Вкладка 2: SADD user:connections:vasya "ws-tab2"
SCARD user:connections:vasya → 2 (два подключения)

Вкладка 1 закрывается: SREM user:connections:vasya "ws-tab1"
SCARD user:connections:vasya → 1 (еще одно подключение)

Вкладка 2 закрывается: SREM user:connections:vasya "ws-tab2"
SCARD user:connections:vasya → 0 (нет подключений)
Статус меняется на "offline"
```

### Сценарий 2: Вася переходит между чатами
```
Был в #general: SET user:active_chat:vasya "general"
Перешел в #random: SET user:active_chat:vasya "random"
GET user:active_chat:vasya → "random" (старый ключ перезаписан)
```

### Сценарий 3: Автоматическая очистка
```
Вася начал печатать: SET user:typing_chat:vasya "chat1" EX 10
Через 10 секунд Redis автоматически удаляет ключ
GET user:typing_chat:vasya → nil (ключ исчез)
Значит Вася перестал печатать
```

## 📊 Структура данных в коде:

### В Redis (быстрая память):
```go
// Ключи
user:status:{userID}        = "online" / "offline"
user:connections:{userID}   = ["ws1", "ws2"]  // Set
user:active_chat:{userID}   = "chatID"
user:typing_chat:{userID}   = "chatID"
chat:online_users:{chatID}  = ["user1", "user2"]  // Set
chat:typing_users:{chatID}  = ["user1"]  // Set
```

### В Go памяти (только для быстрого доступа):
```go
type Hub struct {
    Online map[string]*Client  // Только подключенные WebSocket клиенты
    // Остальное в Redis!
}
```

## 🚀 Преимущества этой архитектуры:

1. **Масштабируемость**: Много серверов → один Redis
2. **Надежность**: Если сервер упал, статусы в Redis остались
3. **Автоочистка**: Redis сам удаляет устаревшие статусы
4. **Быстро**: Все операции за миллисекунды

## 🧪 Как тестировать:

1. **Установи Redis**:
```bash
docker run -p 6379:6379 redis
```

2. **Подключись к Redis CLI**:
```bash
redis-cli
```

3. **Попробуй команды**:
```bash
# Создай пользователя
SET user:status:vasya "online" EX 30

# Посмотри статус
GET user:status:vasya

# Добавь в чат
SADD chat:online_users:general "vasya"

# Посмотри кто в чате
SMEMBERS chat:online_users:general

# Узнай когда удалится
TTL user:status:vasya
```

## ❓ Частые вопросы:

**Q: Почему не хранить все в Go памяти?**
A: Потому что если сервер перезапустится - все статусы пропадут. Redis сохраняет данные.

**Q: Почему два ключа для чата (active_chat и typing_chat)?**
A: Потому что пользователь может сидеть в одном чате (#general), а печатать в другом (#random).

**Q: Что если Redis упадет?**
A: Нужно настроить Redis Cluster ил�� Sentinel для отказоустойчивости.

**Q: Как часто обновлять статус "online"?**
A: Каждые 25 секунд (TTL 30 секунд), чтобы ключ не удалился.

## 📈 Что мониторить:

1. **Количество ключей в Redis**: `redis-cli info keyspace`
2. **Использование памяти**: `redis-cli info memory`
3. **Количество подключений**: `redis-cli info clients`
4. **Задержки операций**: `redis-cli --latency`

## 🎬 Пример клиентского кода:

```javascript
// Когда пользователь заходит в чат
function enterChat(chatId) {
    ws.send(JSON.stringify({
        operation: "switch_chat",
        message: { chat_id: chatId }
    }));
}

// Когда начинает печатать
function startTyping() {
    ws.send(JSON.stringify({
        operation: "typing_start", 
        message: { chat_id: currentChatId }
    }));
    
    // Автоматически остановить через 10 сек
    setTimeout(stopTyping, 10000);
}

// Получить статусы всех в чате
function getChatPresence(chatId) {
    fetch(`/api/v1/status/chat/${chatId}/online`)
    .then(response => response.json())
    .then(users => {
        // Показать кто онлайн
        users.forEach(user => showOnline(user));
    });
}
```

Эта система позволяет точно знать где пользователь, что он делает, и показывать это другим в реальном времени! 🚀