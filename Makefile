download:
	@GOPRIVATE=github.com go get github.com/AlexandrGurkin/tasker/client/api

port-forward:
	@ssh -g -L 0.0.0.0:8078:0.0.0.0:22 -N localhost

cp-exec:
	@cp ./vm_agent_linux ~/fileShare/

build:
	@GOOS=linux go build -o vm_agent_linux

cp-agent:
	@scp -P 8078 aleksandr_gurkin@188.187.1.101:~/fileShare/vm_agent_linux ./

swagger:
	@echo Delete generated files
	@rm -rf restapi/operations restapi/doc.go restapi/embedded_spec.go restapi/server.go models client
	@echo Delete completed
	@echo Code generation
	@docker run --rm -it -e GOPATH=/go -v $$(pwd):/work -w /work quay.io/goswagger/swagger:v0.25.0 generate server --exclude-main -f "./api/swagger.yaml"
	@docker run --rm -t --privileged -e GOPATH=/go -v $$(pwd):/work -w /work quay.io/goswagger/swagger:v0.25.0 generate client -f "./api/swagger.yaml" -c client/api -m client/models
	@echo Generation completed