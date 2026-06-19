include .env
export

.PHONY: dev backend web db prod

dev: db
	npm --prefix web run dev &
	cd backend && wgo run ./cmd/server

backend: db
	cd backend && wgo run ./cmd/server

web:
	npm --prefix web run dev

db:
	docker compose --env-file .env up db -d
	until docker compose exec db pg_isready -U familyboard; do sleep 1; done

prod:
	docker compose --env-file .env up --build -d
