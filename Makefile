.PHONY: migrate-up migrate-down force

DB_URL=postgresql://postgres:postgres@localhost:5432/flashdeal_db?sslmode=disable

migrate-up:
    migrate -path migrations -database "$(DB_URL)" -verbose up

migrate-down:
    migrate -path migrations -database "$(DB_URL)" -verbose down

force:
    migrate -path migrations -database "${DB_URL}" force 1