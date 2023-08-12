# Public variables
OUTPUT_DIR ?= out

# Private variables
obj = r3map-benchmark-direct-mount r3map-benchmark-managed-mount r3map-benchmark-migration
all: $(addprefix build/,$(obj))

# Build
build: $(addprefix build/,$(obj))
$(addprefix build/,$(obj)):
	go build -o $(OUTPUT_DIR)/$(subst build/,,$@) ./cmd/$(subst build/,,$@)

# Integration
integration: integration/direct-mount-file integration/direct-mount-directory integration/managed-mount-file integration/managed-mount-directory

integration/direct-mount-file:
	$(OUTPUT_DIR)/r3map-benchmark-direct-mount --remote-backend=file

integration/direct-mount-directory:
	$(OUTPUT_DIR)/r3map-benchmark-direct-mount --remote-backend=directory --remote-backend-chunking

integration/managed-mount-file:
	$(OUTPUT_DIR)/r3map-benchmark-managed-mount --remote-backend=file

integration/managed-mount-directory:
	$(OUTPUT_DIR)/r3map-benchmark-managed-mount --remote-backend=directory --remote-backend-chunking

# Clean
clean:
	rm -rf out

# Dependencies
depend:
	go generate ./...