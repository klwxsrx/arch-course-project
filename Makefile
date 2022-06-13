.PHONY: clean docker-image/% push-image/%

all: build

build: clean \
	bin/auth bin/catalog bin/cart bin/order bin/payment bin/delivery bin/warehouse \
	docker-image/auth docker-image/catalog docker-image/cart docker-image/order docker-image/payment \
		docker-image/delivery docker-image/warehouse

push: push-image/auth push-image/catalog push-image/cart push-image/order push-image/payment \
	push-image/delivery push-image/warehouse

clean:
	rm -rf bin/*

bin/%:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o ./bin/$(notdir $@) ./cmd/$(notdir $@)

docker-image/%:
	docker build -f docker/$(notdir $@)/Dockerfile -t klwxsrx/arch-course-$(notdir $@)-service .

push-image/%:
	docker push klwxsrx/arch-course-$(notdir $@)-service