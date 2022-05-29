# Процесс проведения платежа

![docs/services.puml](http://www.plantuml.com/plantuml/proxy?fmt=svg&src=https://raw.githubusercontent.com/klwxsrx/arch-course-project/docs/purchase_process.puml)

Сервис `Order` является оркестратором процесса, реализует паттерн Saga.

# Установка

## Добавить аддон ingress minikube

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

## Применить k8s манифесты

```shell
kubectl apply -f ./k8s
```

## Запуск тестов
```shell
newman run --env-var="baseUrl=arch.homework" ./full_case.postman_collection.json
```