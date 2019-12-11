DIRECTORY := $(PWD)
CODE_GENERATOR_IMAGE=quay.io/slok/kube-code-generator:v1.13.5
CODE_GENERATOR_PACKAGE=github.com/spotahome/kooper/test/integration/operator
GROUPS_VERSION="superhero:v1alpha1"

default: generate

.PHONY: generate
generate:
	docker run --rm -it \
	-v $(DIRECTORY):/go/src/$(CODE_GENERATOR_PACKAGE) \
	-e PROJECT_PACKAGE=$(CODE_GENERATOR_PACKAGE) \
	-e CLIENT_GENERATOR_OUT=$(CODE_GENERATOR_PACKAGE)/client/k8s \
	-e APIS_ROOT=$(CODE_GENERATOR_PACKAGE)/apis \
	-e GROUPS_VERSION=$(GROUPS_VERSION) \
	-e GENERATION_TARGETS="deepcopy,client" \
	$(CODE_GENERATOR_IMAGE)