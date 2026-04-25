.PHONY: all build test install clean run help

BINARY_NAME=socks5
INSTALL_DIR=/usr/local/bin
CONFIG_DIR=/etc/go-socks5
INIT_DIR=/etc/init.d
SYSTEMD_DIR=/etc/systemd/system
DEFAULT_DIR=/etc/default

all: build

build:
	go build -o $(BINARY_NAME) .

test:
	go test ./... -v -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@rm coverage.out

clean:
	rm -f $(BINARY_NAME) coverage.html
	rm -f *.db *.db-journal

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Creating config directory at $(CONFIG_DIR)..."
	mkdir -p $(CONFIG_DIR)
	install -m 644 config.example.yaml $(CONFIG_DIR)/config.yaml
	install -m 644 users.schema.sql $(CONFIG_DIR)/users.schema.sql
	install -m 644 traffic.schema.sql $(CONFIG_DIR)/traffic.schema.sql
	install -m 644 socks5.default $(DEFAULT_DIR)/$(BINARY_NAME)
	@echo "Installing systemd service..."
	install -m 644 $(BINARY_NAME).service $(SYSTEMD_DIR)/$(BINARY_NAME).service
	systemctl daemon-reload
	@echo "Installing SysV init script..."
	install -m 755 socks5.init $(INIT_DIR)/$(BINARY_NAME)
	update-rc.d $(BINARY_NAME) defaults 2>/dev/null || true
	@echo ""
	@echo "Installation complete!"
	@echo ""
	@echo "For systemd systems:"
	@echo "  systemctl enable $(BINARY_NAME)  # enable on boot"
	@echo "  systemctl start $(BINARY_NAME)   # start now"
	@echo ""
	@echo "For SysV systems:"
	@echo "  update-rc.d $(BINARY_NAME) defaults"
	@echo "  /etc/init.d/$(BINARY_NAME) start"

uninstall:
	@echo "Stopping service..."
	-systemctl stop $(BINARY_NAME) 2>/dev/null || /etc/init.d/$(BINARY_NAME) stop 2>/dev/null || true
	@echo "Disabling service..."
	-systemctl disable $(BINARY_NAME) 2>/dev/null || true
	update-rc.d $(BINARY_NAME) remove 2>/dev/null || true
	@echo "Removing systemd service..."
	rm -f $(SYSTEMD_DIR)/$(BINARY_NAME).service
	systemctl daemon-reload
	@echo "Removing SysV init script..."
	rm -f $(INIT_DIR)/$(BINARY_NAME)
	@echo "Removing binary..."
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Config directory preserved at $(CONFIG_DIR)"
	@echo "Default config preserved at $(DEFAULT_DIR)/$(BINARY_NAME)"

run: build
	./$(BINARY_NAME) -v

help:
	@echo "Available targets:"
	@echo "  make build        - Build the application"
	@echo "  make test        - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make install    - Install binary, config, and service files"
	@echo "  make uninstall  - Remove installed components"
	@echo "  make run        - Build and run locally"
