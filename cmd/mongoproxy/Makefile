test:
	CGO_ENABLED=0 go build -mod=vendor
	SKIP_RECOVER=true ./mongoproxy --config=test.conf --log-level=debug --metrics-bind=localhost:8888 --term-sleep=0s

ci:
	CGO_ENABLED=0 go build -mod=vendor
	SKIP_RECOVER=true ./mongoproxy --config=ci.conf --log-level=debug --metrics-bind=localhost:8888 --term-sleep=0s
