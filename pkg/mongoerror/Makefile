default:
	python errcodegen.py ~/src/mongo/src/mongo/base/error_codes.err > errorgen.go
	goimports -w errorgen.go
