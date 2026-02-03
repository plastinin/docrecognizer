# docrecognizer

## Шаг 1:  Настройки .env запуск инфраструктуру

```
cp .env.example .env #можно скорректировать OLLAMA_MODEL=
make infra-up
```

## Шаг 2: Скачайте модель Ollama

```
make ollama-pull
docker logs -f docrecognizer-ollama
```

## Шаг 3: Миграции 

```
make migrate-up
```

## Шаг 4: API Run

```
make run-api
```

## Шаг 5: Запустите Worker (во втором терминале)

```
make run-worker
```

## Шаг 6: health check

```
curl http://localhost:8080/health
{"status":"ok"}
```