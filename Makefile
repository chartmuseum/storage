.PHONY: clean
clean:
	git status --ignored --short | grep '^!! ' | sed 's/!! //' | xargs rm -rf

.PHONY: covhtml
covhtml:
	open coverage.html

.PHONY: test
test:
	rm -rf .test/ && mkdir .test/
	go test -v -covermode=atomic -coverprofile=coverage.out .
	go tool cover -html=coverage.out -o=coverage.html

.PHONY: testcloud
testcloud: export TEST_CLOUD_STORAGE=1
testcloud: test
