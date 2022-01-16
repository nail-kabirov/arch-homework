# Описание
#### Домашнее задание выполнено для курса "[Микросервисная архитектура](https://otus.ru/lessons/microservice-architecture)"

# Сервисы
Сами сервисы расположен в папке `services`, при запуске в ней команды `make` собираются docker-образы сервисов.
API всех сервисов можно посмотреть в папке `service/api`

# Инструкция по запуску

### Предварительная подготовка
1. Прописать в hosts домен `arch.homework` на ip кластера
2. При необходимости создать новый namespace и выбрать его, например:
```
kubectl create namespace arch-hw8 && kubectl config set-context --current --namespace=arch-hw8
```
3. Установить Nginx при отсутствии (или включить addon в minikube - `minikube addons enable ingress`)

### Установка приложения с помощью helm:
```
helm install hw8 helm/hw-umbrella-chart
```
# Тестирование
### Запуск тестов:
```
 newman run tests.postman_collection.json
```