This directory structure exists to provide stub package implementations for the linter tests. The `analysistest` package replaces `$GOPATH` with the local file system path. We create these stubs so the linter test files can closely mimic our production code.