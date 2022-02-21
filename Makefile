PACKAGE=nt-folly-xmaxx-comp

LOCAL_DEV_DB_USERNAME=$(PACKAGE)
LOCAL_DEV_DB_PASS=dev
LOCAL_DEV_DB_HOST=localhost
LOCAL_DEV_DB_PORT=5444
LOCAL_DEV_DB_DATABASE=$(PACKAGE)
DB_CONNECTION_STRING="postgres://$(LOCAL_DEV_DB_USERNAME):$(LOCAL_DEV_DB_PASS)@$(LOCAL_DEV_DB_HOST):$(LOCAL_DEV_DB_PORT)/$(LOCAL_DEV_DB_DATABASE)?sslmode=disable"

.PHONY: build
build:
	@mkdir -p dist
	go build -o dist/migrate cmd/migrate/main.go
	go build -o dist/collection cmd/collection/main.go
	go build -o dist/serve cmd/serve/main.go

.PHONY: db-start
db-start:
	$(CURDIR)/scripts/docker-start-localdb.sh $(PACKAGE) $(LOCAL_DEV_DB_PORT)

.PHONY: db-stop
db-stop:
	docker stop $(PACKAGE)-db	

.PHONY: db-remove
db-remove:
	docker rm $(PACKAGE)-db	

.PHONY: db-migrate-drop
db-migrate-drop:
	go run cmd/migrate/main.go migrate-drop

.PHONY: db-migrate-up
db-migrate-up:
	go run cmd/migrate/main.go migrate-up

.PHONY: db-migrate-up-1
db-migrate-up-1:
	go run cmd/migrate/main.go migrate-up 1

.PHONY: db-migrate-down
db-migrate-down:
	go run cmd/migrate/main.go migrate-down

.PHONY: db-migrate-down-1
db-migrate-down-1:
	go run cmd/migrate/main.go migrate-down 1

.PHONY: db-migrate-repeat-1
db-migrate-repeat-1: db-migrate-down-1 db-migrate-up-1

.PHONY: db-reset
db-reset: db-migrate-drop db-migrate-up

.PHONY: gql
gql:
	go run github.com/99designs/gqlgen generate

.PHONY: collection
collection:
	go run cmd/collection/main.go service

.PHONY: serve
serve:
	go run cmd/serve/main.go service

