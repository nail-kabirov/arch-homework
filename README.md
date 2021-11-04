# Описание
#### Домашнее задание выполнено для курса "[Микросервисная архитектура](https://otus.ru/lessons/microservice-architecture)"

# Сервис
Сам сервис расположен в папке `service`, при запуске в ней команды `make` собирается docker-образ с сервисом.

# Инструкция по запуску

### Предварительная подготовка
1. Прописать в hosts домен `arch.homework` на ip кластера
2. При необходимости создать новый namespace и выбрать его, например:
```
kubectl create namespace arch-hw2 && kubectl config set-context --current --namespace=arch-hw2
```
3. Добавить в helm репозиторий `bitmani`, если он еще был добавлен ранее:
```
helm repo add bitnami https://charts.bitnami.com/bitnami
```

### Установка Postgres из helm:
```
helm install postgresql -f k8s/postgres/postgresql-values.yaml bitnami/postgresql
```

### Применение первоначальной миграции к базе: 
```
kubectl apply -f k8s/initial_migration.job.yaml
```
### Применение манифестов для развертывания сервиса в кластере
```
kubectl apply -f k8s
```
# Тестирование сервиса
В файле `user.postman_collection.json` представлена postman-коллекция с запросами к сервису.
Можно убедиться в работоспособности вызовом:
```
newman run user.postman_collection.json
```
