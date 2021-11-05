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
3. Добавить `role` и `rolebinding` в текущем namespace'е для возможности получать список job'ов внутри кластера (необходимо для корректной работы initContainer'а в deployment'е):
```
export CURRENT_NAMESPACE=`kubectl config view --minify --output 'jsonpath={..namespace}'`
kubectl create role job-reader --verb=get --verb=list --verb=watch --resource=jobs
kubectl create rolebinding $CURRENT_NAMESPACE-job-reader --role=job-reader --serviceaccount=$CURRENT_NAMESPACE:default
```
4. Добавить в helm репозиторий `bitmani`, если он еще был добавлен ранее:
```
helm repo add bitnami https://charts.bitnami.com/bitnami
```

### Установка Postgres из helm:
```
helm install postgresql -f k8s/postgres/postgresql-values.yaml bitnami/postgresql
```
### Применение манифестов для развертывания сервиса в кластере (вместе с миграцией, завершения которой будет ждать deployment)
```
kubectl apply -f k8s
```

При этом можно даже эту команду запустить одновременно с установкой postgres. В этом случае скорее всего миграция с первого раза не выполнится, но выполнится при одном из перезапусков (ограничение стоит на 10 с таймаутом на выполнение 5 секунд).
Deployment сервиса же будет ждать окончания выполнения миграции с помощью initContainer'а `groundnuty/k8s-wait-for`. Переданный ему параметр `job-wr` означает, что нужно ждать выполнения job'а, при этом необходимо хотя бы одно успешное завершение работы, а количество ошибочных не важно.

# Тестирование сервиса
В файле `user.postman_collection.json` представлена postman-коллекция с запросами к сервису.
Можно убедиться в работоспособности вызовом:
```
newman run user.postman_collection.json
```
