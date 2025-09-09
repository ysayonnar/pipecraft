build_alpine_git:
	docker build -t dind-git -f ../services/dind-git.dockerfile .

up:
	$(MAKE) build_alpine_git
	docker compose up --build -d

up_logs:
	$(MAKE) build_alpine_git
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

restart:
	docker compose down && docker compose up --build -d

restart_logs:
	docker compose down && docker compose up --build

test:
	go test -C ../services/internal ./handlers
