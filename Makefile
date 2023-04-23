.PHONY: dummy-service
dummy-service:
	go run ./test/dummy_service/dummy_service.go -port 8081 &
	go run ./test/dummy_service/dummy_service.go -port 8082 &
	sleep 2

.PHONY: kill-dummy
kill-dummy:
	sudo kill -9 $$(sudo lsof -t -i:8081)
	sudo kill -9 $$(sudo lsof -t -i:8082)

.PHONY: test
test:
	go test -v ./test/e2e/