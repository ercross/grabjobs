# Makefile for development
BINARY_DIR=bin
BINARY_NAME=grabjobs
PORT=4045
LOCATION_DATA=res/location_data.csv

clean_binary:
	@echo "cleaning api binary files..."
	@go clean
	@- rm -f ${BINARY_DIR}/${BINARY_NAME}
	@echo "api binary files cleaned"

build_binary:
	@echo "building Grabjobs api..."
	@go build -o ${BINARY_DIR}/${BINARY_NAME} ./cmd/api/*.go
	@echo "Grabjobs api built"

start_app:
	./${BINARY_DIR}/${BINARY_NAME} -port ${PORT} -db ${LOCATION_DATA}

run_test:
	 go test -v ./internal/db

run_app: clean_binary build_binary start_app

stop_app:
	@echo "stopping app: " ${BINARY_NAME}
	killall -e ${BINARY_NAME}

restart_app: stop_app run_app

