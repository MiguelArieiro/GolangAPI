.PHONY: docker-up
docker-up: ## Run app and DB on container
	docker-compose -f docker-compose.yaml up --build

.PHONY: docker-down
docker-down: ## Stop docker containers and clear artefacts.
	docker-compose -f docker-compose.yaml down
	docker-compose -f docker-compose-test.yaml down
	docker system prune --volumes

.PHONY: docker-test
docker-test: ## Run tests and DB on container
	docker-compose -f docker-compose-test.yaml up --build --abort-on-container-exit