🍕 Pizza Shop REST API

Бэкенд-сервис для оформления заказов пиццы, написанный на Go с использованием фреймворка Gin и СУБД PostgreSQL.

Проект предоставляет полноценный REST API с ролевой моделью доступа (пользователь / администратор), безопасной аутентификацией через JWT и обработкой заказов в транзакционной среде PostgreSQL.

---

## 🛠 Технологический стек

* **Язык программирования:** Go (1.21+)
* **Веб-фреймворк:** [Gin Web Framework](https://github.com/gin-gonic/gin)
* **База данных:** PostgreSQL
* **Драйвер БД:** [pgx/v5](https://github.com/jackc/pgx) (работа с пулом соединений `pgxpool`)
* **Аутентификация:** JWT (JSON Web Tokens)
* **Хеширование паролей:** `golang.org/x/crypto/bcrypt`
* **Конфигурация:** `godotenv`

Создайте файл .env в корневой директории проекта и укажите ваши параметры подключения:
DATABASE_URL=postgres://postgres:password@localhost:5432/pizza_db?sslmode=disable
SERVER_PORT=:8080
JWT_SECRET=super_secret_jwt_key_change_me
