.PHONY: lint gen update-runners
lint:
	golangci-lint fmt
	golangci-lint run -c .golangci.yml

gen:
	sqlc generate -f ./sql/sqlc.yml

update-runners:
	docker run --privileged --rm tonistiigi/binfmt --install all
	docker buildx build -f docker/Dockerfile.callback --platform linux/amd64 -t yagatito/callback-service-prod:1.0  --push ./
	docker buildx build -f docker/Dockerfile.wallet		--platform linux/amd64 -t yagatito/wallet-service-prod:1.0 		--push ./
	docker buildx build -f docker/Dockerfile.worker		--platform linux/amd64 -t yagatito/worker-service-prod:1.0 		--push ./

go-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
# 	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
