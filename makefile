.PHONY: build

build:
	go build

postgres:
	$(MAKE) install-postgres
	@/bin/sleep 10
	$(MAKE) config-postgres

install-postgres:
	@docker run \
		--name antidose-pg \
		-e POSTGRESQL_USER=anti \
		-e POSTGRESQL_PASSWORD=naloxone499 \
		-d -p 5432:5432 mdillon/postgis
	
config-postgres:
	@-psql -h localhost -p 5432 -U postgres -f db/db_init.sql	


clean:
	@docker rm -f antidose-pg
	@echo "Postgres Container Removed"
