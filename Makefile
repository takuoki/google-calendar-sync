#: generate open API files
.PHONY: gen-openapi
gen-openapi:
	@rm -f ./api/openapi/openapi.gen.go
	@oapi-codegen ./api/openapi/openapi.yml > ./api/openapi/openapi.gen.go

#: run API locally
.PHONY: run
run:
	@docker-compose up --build

#: down local containers
.PHONY: down
down:
	@docker-compose down
