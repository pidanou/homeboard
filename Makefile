include .env
export

.PHONY: dev backend web db prod ios-dev android-dev

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

ios-dev:
	@cd web && LOCAL_IP=$$(ipconfig getifaddr en0) && \
	LIVE_RELOAD_URL="http://$$LOCAL_IP:5173" npm exec -- cap sync ios && \
	npm run dev -- --host 0.0.0.0 & \
	npm exec -- cap open ios

android-dev:
	@cd web && LOCAL_IP=$$(ipconfig getifaddr en0) && \
	LIVE_RELOAD_URL="http://$$LOCAL_IP:5173" npm exec -- cap sync android && \
	npm run dev -- --host 0.0.0.0 & \
	npm exec -- cap open android
