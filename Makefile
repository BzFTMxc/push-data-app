run: bin/push-data-app
	@PATH="$(PWD)/bin:$(PATH)" heroku local

bin/push-data-app: mock.go
	go build -o bin/push-data-app mock.go

clean:
	rm -rf bin
