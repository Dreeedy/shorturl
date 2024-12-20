## Форматирование
Форматирование всех .go файлов
go fmt ./...

make golangci-lint-clean
make golangci-lint-run

Актуализация всех зависимостей
go mod tidy
## Тестирование

Для выполнения всех тестов в проекте используйте следующую команду:
go test ./... -v

## Сборка приложения
Для сборки приложения выполните следующую команду:
go build -o shortener

В данной директории будет содержаться код, который скомпилируется в бинарное приложение:
cmd\shortener\

## Запуск приложения

### С переменными окружения
Установите переменные окружения и запустите приложение:
\$env\:SERVER_ADDRESS=":8081"; \$env\:BASE_URL="http://localhost:8081"; ./shortener.exe

### С флагами командной строки
Запустите приложение с флагами командной строки:
./shortener -a :8888 -b http://localhost:8888

### С переменными окружения и флагами командной строки
Установите переменные окружения и запустите приложение с флагами командной строки (переменные окружения имеют приоритет):
\$env\:SERVER_ADDRESS=":8081"; \$env\:BASE_URL="http://localhost:8081"; ./shortener -a :8888 -b http://localhost:8888

## Работа с моками

### Пример создание мока
& "C:\Program Files\Go\bin\bin\mockgen.exe" -source="F:\shorturl\internal\storage\storage.go" -destination="F:\shorturl\internal\storage\storage_mock.go" -package=storage