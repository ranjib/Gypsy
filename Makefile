
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
				 -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

bin:
	go build -o gypsy main.go version.go commands.go

deps:
	go get -d -v ./...

format:
	go fmt ./...

vet:
	@go tool vet $(VETARGS) . ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for reviewal."; \
	fi

.PHONY: bin deps format vet
