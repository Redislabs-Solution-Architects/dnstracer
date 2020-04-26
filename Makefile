TEST?=.
NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' dnstracer.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' dnstracer.go)


all: test xcompile

test:
	go test $(TEST) -v
	go vet $(TEST)

xcompile:
	@rm -rf build/
	@mkdir -p build
	gox \
		-os="darwin" \
		-os="linux" \
		-output="build/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}/$(NAME)"
