# Public variables
OUTPUT_DIR ?= out

# Private variables
obj = r3map-benchmark-direct-mount r3map-benchmark-managed-mount r3map-benchmark-migration
all: $(addprefix build/,$(obj))

# Build
build: $(addprefix build/,$(obj))
$(addprefix build/,$(obj)):
	go build -o $(OUTPUT_DIR)/$(subst build/,,$@) ./cmd/$(subst build/,,$@)

# Test
test:
	go test -timeout 3600s -parallel $(shell nproc) ./...

# Integration
integration: integration/direct-mount-file integration/direct-mount-directory integration/managed-mount-file

integration/direct-mount-file:
	$(OUTPUT_DIR)/r3map-benchmark-direct-mount --remote-backend=file

integration/direct-mount-directory:
	$(OUTPUT_DIR)/r3map-benchmark-direct-mount --remote-backend=directory --remote-backend-chunking

integration/managed-mount-file:
	$(OUTPUT_DIR)/r3map-benchmark-managed-mount --remote-backend=file

# Clean
clean:
	rm -rf out

# Dependencies
depend:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/loopholelabs/frpc-go/protoc-gen-go-frpc@latest

	go generate ./...