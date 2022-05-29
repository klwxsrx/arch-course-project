# Техническое решение

## Карта сервисов

![docs/services.puml](http://www.plantuml.com/plantuml/proxy?fmt=svg&src=https://raw.githubusercontent.com/klwxsrx/arch-course-project/master/docs/services.puml)

Интернет-магазин распилен на 7 микросервисов, каждый сервис реализует свой небольшой домен (ограниченный контекст в понятиях DDD).

С целью экономии времени на разработку общение между сервисами происходит синхронно, в идеале стоит реализовать асинхронное взаимодействие используя очередь событий на базе `RabbitMQ`/`Kafka`/... 



## Аутентификация и авторизация

Каждый запрос в сервис попадает в `API Gateway`, где осуществляется аутентификация пользователя.

В качестве `API Gateway` в проекте используется `Traefik`.

### Фронтендовая аутентификация

![docs/user_auth.puml](http://www.plantuml.com/plantuml/proxy?fmt=svg&src=https://raw.githubusercontent.com/klwxsrx/arch-course-project/master/docs/user_auth.puml)

Запросы, используемые фронтендом и непосредственно пользователями интернет-магазина имеют префикс `/web/{serviceName}/`.
Для большинства из них требуется залогиненный пользователь.

Для кейса фронтендовой аутентификации используется `forwardAuth` в `Traefik`.

Сервис `Auth` отвечает за регистрацию и авторизацию пользователей.
Запросы по префиксу `/auth/*` форвардятся `Traefik`'ом в него.

При успешной логинации в `Auth` пользователю проставляется сессионная кука `sid` (рандомная строка).

На каждый фронтовый запрос требующий аутентификации `Traefik` обращается к сервису `Auth`, где происходит сопоставление пользователя по сессионной куке.

Далее в сервис передается заголовок залогиненного пользователя `X-Auth-User-ID` в запросе.

### Внутренняя аутентификация

Для внутренних целей, например добавления продукта в каталог магазина, добавления товара на склад и т.п. используется `basicAuth`.

В запросы по префиксу `/{serviceName}/` требуется проставлять заголовок `Authorization: Basic dXNlcjoxMjM0` (пользователь и пароль `user:1234`).

См. [postman-коллекцию тестов](https://github.com/klwxsrx/arch-course-project#запуск-тестов) ниже.

## Процесс проведения платежа

![docs/purchase_process.puml](http://www.plantuml.com/plantuml/proxy?fmt=svg&src=https://raw.githubusercontent.com/klwxsrx/arch-course-project/master/docs/purchase_process.puml)

Сервис `Order` является оркестратором процесса, реализует паттерн Saga. В случае провала на каком-либо шаге все предыдущие действия откатятся.

# Установка

## Добавить addon ingress minikube

```shell
minikube addons enable ingress
```

## Установить Traefik

```shell
helm repo add traefik https://helm.traefik.io/traefik
helm repo update
helm install --version "10.14.1" \
  --set ports.websecure.expose=false \
  --set providers.kubernetesIngress.enabled=false \
  --set providers.kubernetesCRD.namespaces=arch-course \
  traefik traefik/traefik --namespace traefik --create-namespace
```

## Развернуть MySQL

```shell
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm install \
  --set "image.tag=8.0" \
  --set "auth.database=archcourse" \
  --set "auth.username=user" \
  --set "auth.password=test1234" \
  --namespace=arch-course \
  --create-namespace \
  arch-course-db bitnami/mysql
```

## Развернуть Redis

```shell
helm install \
  --set "auth.enabled=true" \
  --set "auth.password=test1234" \
  --set "replica.replicaCount=0" \
  --namespace=arch-course \
  --create-namespace \
  arch-course-redis bitnami/redis
```

## Применить K8S манифесты

```shell
kubectl apply -f ./k8s
```

# Запуск тестов

```shell
newman run --env-var="baseUrl=arch.homework" ./full_case.postman_collection.json
```