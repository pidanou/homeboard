include .env
export

.PHONY: dev backend web db prod ios-dev android-dev ios-open android-open

dev: db
	npm --prefix web run dev &
	cd backend && wgo run ./cmd/server

backend: db
	cd backend && wgo run ./cmd/server

web:
	LOCAL_IP=$$(ipconfig getifaddr en0) npm --prefix web run dev -- --host 0.0.0.0

db:
	docker compose --env-file .env up db -d
	until docker compose exec db pg_isready -U familyboard; do sleep 1; done

prod:
	docker compose --env-file .env up --build -d

ios-dev:
	cd web && LIVE_RELOAD_URL="http://$$(ipconfig getifaddr en0):5173" npm exec -- cap sync ios

android-dev:
	cd web && LIVE_RELOAD_URL="http://$$(ipconfig getifaddr en0):5173" npm exec -- cap sync android

ios-open:
	cd web && npm exec -- cap open ios

android-open:
	cd web && npm exec -- cap open android
