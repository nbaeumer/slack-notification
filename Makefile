.PHONY: deps clean build deploy

deps:
	go get -u ./...

clean:
	rm -rf dist/slack-notification

build:
	GOOS=linux GOARCH=amd64 go build -o dist/slack-notification main.go

run:
	SLACK_WEBHOOK="AQICAHitEE9ZBAkn4k0Nyux9EfWlAGAryjGrN8cG3sZY4YU0UAHV91iHoWZPThy8gkL8IdTmAAAAizCBiAYJKoZIhvcNAQcGoHsweQIBADB0BgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDOesTWq+3xL+03lW/AIBEIBH+rGs1b74Madi29kG8FO6IBZs3VqGQ+m173BX8+15xrbS7KI+hlX9RxKkhsCIAMBkvy9E+O1GqK0GIuWztIvklEe+HCNa8Q8=" LOG_GROUP="ng-docs" KmsKeyId="ce17fa91-7b37-4c2c-a98a-d939cf48370f" sam local invoke --template template.yaml --event event.json
deploy:
	sh deploy.sh