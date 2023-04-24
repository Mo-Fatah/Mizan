.PHONY: example-service
example-service:
	go run ./examples/example_service.go -port 9090 &
	go run ./examples/example_service.go -port 9091 &
	go run ./examples/example_service.go -port 9092 &
	sleep 2

.PHONY: kill-example
kill-example:
	sudo kill -9 $$(sudo lsof -t -i:9090)
	sudo kill -9 $$(sudo lsof -t -i:9091)
	sudo kill -9 $$(sudo lsof -t -i:9092)

.PHONY: test-e2e
test-e2e: 
	go test ./test/e2e/

.PHONY: test-unit
test-unit:
	go test ./internal/...

.PHONY: test
test: test-unit test-e2e
