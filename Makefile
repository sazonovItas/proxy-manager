.PHONY: lint
lint:
	golangci-lint run --config .golangci.yml ./pkg/... ./services/proxy/... ./services/proxy-manager/...
