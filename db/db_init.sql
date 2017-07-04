CREATE TYPE user_status AS ENUM ('active', 'inactive', 'deleted', 'unverified');

CREATE TABLE IF NOT EXISTS users (
	u_id                serial PRIMARY KEY,
	first_name          VARCHAR(20),
	last_name           VARCHAR(20),
	phone_number        VARCHAR(16) UNIQUE,
	current_status      user_status,
	token				VARCHAR(6)
);

CREATE TABLE IF NOT EXISTS temp_users (
	temp_u_id			serial,
	first_name			VARCHAR(20),
	last_name			VARCHAR(20),
	phone_number		VARCHAR(16) PRIMARY KEY,
	token				VARCHAR(6),
	init_time			TIMESTAMP WITHOUT TIME ZONE
);

CREATE TABLE IF NOT EXISTS incidents (
	inc_id              serial PRIMARY KEY,
	requester_imei      VARCHAR(15),
	req_by_helper       BOOLEAN,
	time_start          TIMESTAMP WITHOUT TIME ZONE,
	time_end            TIMESTAMP WITHOUT TIME ZONE
);

SELECT AddGeometryColumn('incidents', 'init_req_location', 4326, 'POINT', 2);

CREATE TABLE IF NOT EXISTS requests (
	req_id              serial PRIMARY KEY,
	u_id                INTEGER REFERENCES users(u_id),
	init_time           TIMESTAMP WITHOUT TIME ZONE,
	time_reponded       TIMESTAMP WITHOUT TIME ZONE,
	reponse_val         BOOLEAN,
	inc_id              INTEGER REFERENCES incidents(inc_id)
);

--  Adds a 2 dimensional Geometry column of type point, in SRID 4326
SELECT AddGeometryColumn('requests', 'init_help_location', 4326, 'POINT', 2);


CREATE TABLE IF NOT EXISTS location (
	u_id                INTEGER REFERENCES users(u_id)
);

SELECT AddGeometryColumn('location', 'help_location', 4326, 'POINT', 2);
