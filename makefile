server:
	go run main.go

test:
	go test tests/* -v

.PHONY: clean gen server client test cert 
