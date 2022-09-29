.PHONY: build clean

etcd:
	@bash ./scripts/etcd2.sh

build:
	@echo "build start"
	@cd cmd/apiserver && go build -o apiserver && cd ..
	@cd cmd/kubectl && go build -o kubectl && cd ..
	@cd cmd/controller-manager && go build -o controller-manager && cd ..
	@cd cmd/integration && go build -o integration && cd ..
	@cd cmd/kubelet && go build -o kubelet && cd ..
	@cd cmd/proxy && go build -o proxy && cd ..
	@echo "build done"

clean:
	@rm cmd/apiserver/apiserver || true
	@rm cmd/kubectl/kubectl
	@rm cmd/controller-manager/controller-manager || true
	@rm cmd/integration/integration || true
	@rm cmd/kubelet/kubelet || true
	@rm cmd/proxy/proxy || true

integration-test: build
	@cd cmd/integration && ./integration