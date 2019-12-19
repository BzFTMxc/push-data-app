run: bin/push-data-app
	@PATH="$(PWD)/bin:$(PATH)" heroku local

bin/push-data-app: main.go
	go build -o bin/push-data-app main.go

clean:
	rm -rf bin
