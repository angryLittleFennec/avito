all: run

.PHONY: run test stop test-stop rebuild test-rebuild

run:
	docker-compose up --remove-orphans --force-recreate

test:
	docker-compose -f docker-compose-test.yml down --volumes
	docker-compose -f docker-compose-test.yml up  --build --remove-orphans --force-recreate e2e-tests

stop:
	docker-compose down

test-stop:
	docker-compose -f docker-compose-test.yml down --volumes

rebuild:
	docker-compose up --build --remove-orphans --force-recreate

