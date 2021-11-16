# Описание
#### Домашнее задание выполнено для курса "[Микросервисная архитектура](https://otus.ru/lessons/microservice-architecture)"

# Сервис
Сам сервис расположен в папке `service`, при запуске в ней команды `make` собирается docker-образ с сервисом.
API сервиса можно посмотреть в файле `service/api/userapi.swagger.yaml`

# Инструкция по запуску

### Предварительная подготовка
1. Прописать в hosts домен `arch.homework` на ip кластера
2. При необходимости создать новый namespace и выбрать его, например:
```
kubectl create namespace arch-hw3 && kubectl config set-context --current --namespace=arch-hw3
```
3. Установить Prometheus и Nginx при отсутствии:
```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add ingress-nginx 
helm repo update

helm install prom prometheus-community/kube-prometheus-stack -f external/prometheus.yaml --atomic
helm install nginx ingress-nginx/ingress-nginx -f external/nginx-ingress.yaml --atomic
``` 

### Установка приложения с помощью helm:
```
helm install user-app helm/user-app-chart
```

При запуске первоначальная миграция скорее всего с первого раза не выполнится (так как сам postgres не сразу становится доступным), но выполнится при одном из перезапусков (ограничение стоит на 10 перезапусков с таймаутом на выполнение 5 секунд).
Deployment сервиса же будет ждать окончания выполнения миграции с помощью initContainer'а `groundnuty/k8s-wait-for`. Переданный ему параметр `job-wr` означает, что нужно ждать выполнения job'а, при этом необходимо хотя бы одно успешное завершение работы, а количество ошибочных не важно.
Также для корректной работы мониторинга выполнения job'а миграции в чарт включены кастомные `ServiceAccount`, `Role` и `RoleBinding`.

# Тестирование сервиса
Запустить стресс-тест сервиса:
```
external/stress.sh
```

Теперь можно открыть графану по адресу [http://localhost:9000](http://localhost:9000) прокинув порт к ней:
```
kubectl   port-forward service/prom-grafana 9000:80
```
В графане должен появиться dashboard `User-app`
