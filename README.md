# db-forum
API к базе данных проекта "Форумы"

## Документация к API
Документация к API представлена в виде спецификации OpenAPI: swagger.json

Документацию можно читать через Swagger UI: https://tech-db-forum.bozaro.ru/

## Запуск контейнера
Контейнер запускается командами вида:
```
docker build -t forum https://github.com/maratishimbaev/db-forum.git
docker run -p 5000:5000 --name forum -t forum
```