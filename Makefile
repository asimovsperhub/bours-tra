# protoc build
.PHONY: protoc
protoc:
	protoc --go_out=. pkg/protocol/*.proto