PKGS := pkg/client

BUILD_PKGS = $(foreach pkg,$(PKGS),build-$(pkg))
TEST_PKGS = $(foreach pkg,$(PKGS),test-$(pkg))

build: $(BUILD_PKGS)
.PHONY: build

$(BUILD_PKGS): build-%:
	cd $* && go build
.PHONY: $(BUILD_PKGS)

test: $(TEST_PKGS)
.PHONY: test

$(TEST_PKGS): test-%:
	cd $* && go test
.PHONY: $(TEST_PKGS)
