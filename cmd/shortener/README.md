## Форматирование
Форматирование всех .go файлов
go fmt ./...

make golangci-lint-clean
make golangci-lint-run

Актуализация всех зависимостей
go mod tidy

Vet изучает исходный код Go и сообщает о подозрительных конструкциях, таких как вызовы Printf, аргументы которых не совпадают со строкой формата.
go vet ./...

Проверяет синтаксис тегов JSON.
go vet -structtag ./...

gzip -c data.json > data.json.gz
-- Создает сжатый архив правильного формата. Далее заливаем в постман.
## Тестирование

Для выполнения всех тестов в проекте используйте следующую команду:
go test ./... -v

## Сборка приложения
Для сборки приложения выполните следующую команду:
go build -o shortener

В данной директории будет содержаться код, который скомпилируется в бинарное приложение:
cmd\shortener\

Пример установки пакета
go get github.com/jmoiron/sqlx

## Запуск приложения

### С переменными окружения
Установите переменные окружения и запустите приложение:
\$env\:SERVER_ADDRESS=":8081"; \$env\:BASE_URL="http://localhost:8081"; ./shortener.exe

### С флагами командной строки
Запустите приложение с флагами командной строки:
./shortener -a :8888 -b http://localhost:8888
go run main.go -d "user=postgres dbname=mydb sslmode=disable password=111 host=localhost port=5432"

### С переменными окружения и флагами командной строки
Установите переменные окружения и запустите приложение с флагами командной строки (переменные окружения имеют приоритет):
\$env\:SERVER_ADDRESS=":8081"; \$env\:BASE_URL="http://localhost:8081"; ./shortener -a :8888 -b http://localhost:8888

Default local DBConnectionAdress:
"user=postgres dbname=mydb sslmode=disable password=111 host=localhost port=5432"

## Работа с моками

### Пример создание мока
& "C:\Program Files\Go\bin\bin\mockgen.exe" -source="F:\shorturl\internal\storage\storage.go" -destination="F:\shorturl\internal\storage\storage_mock.go" -package=storage