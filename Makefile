# Load environment variables from .env file
-include .env
export


#: generate open API files
.PHONY: gen-openapi
gen-openapi:
	@rm -f ./api/openapi/openapi.gen.go
	@oapi-codegen ./api/openapi/openapi.yml > ./api/openapi/openapi.gen.go

#: run golangci-lint
.PHONY: lint
lint:
	@golangci-lint run ./api/...

#: run API locally
.PHONY: run
run:
	@docker-compose up --build

#: down local containers
.PHONY: down
down:
	@docker-compose down

#: build and push image to artifact registry
.PHONY: build
build:
	@gcloud builds submit --tag $(IMAGE_NAME) api/

#: deploy API to cloud run
.PHONY: deploy
deploy:
	@gcloud run deploy $(SERVICE_NAME) \
		--image $(IMAGE_NAME) \
		--add-cloudsql-instances $(PROJECT_ID):$(REGION):$(INSTANCE_NAME) \
		--region $(REGION) \
		--platform managed \
		--allow-unauthenticated \
		--service-account $(SERVICE_ACCOUNT) \
		--set-env-vars LOG_LEVEL=Debug \
		--set-env-vars DB_TYPE=cloudsql \
		--set-env-vars INSTANCE_CONNECTION_NAME=$(PROJECT_ID):$(REGION):$(INSTANCE_NAME) \
		--set-env-vars DB_NAME=$(DB_NAME) \
		--set-env-vars DB_USER=$(DB_USER) \
		--update-secrets DB_PASSWORD=$(DB_PASSWORD_SECRET) \
		--set-env-vars WEBHOOK_BASE_URL=$(API_URL)/api/sync \
		$(if $(OAUTH_CLIENT_ID),--set-env-vars OAUTH_CLIENT_ID=$(OAUTH_CLIENT_ID)) \
		$(if $(OAUTH_CLIENT_SECRET),--update-secrets OAUTH_CLIENT_SECRET=$(OAUTH_CLIENT_SECRET)) \
		$(if $(OAUTH_REDIRECT_URL),--set-env-vars OAUTH_REDIRECT_URL=$(OAUTH_REDIRECT_URL)) \
		$(if $(CRYPT_KEY_SECRET),--update-secrets CRYPT_KEY=$(CRYPT_KEY_SECRET))
