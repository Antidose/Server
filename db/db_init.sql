CREATE USER anti;
ALTER USER anti WITH PASSWORD 'naloxone499';

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'deleted', 'unverified');

CREATE TABLE IF NOT EXISTS users (
	u_id                serial PRIMARY KEY,
	first_name          VARCHAR(20),
	last_name           VARCHAR(20),
	phone_number        VARCHAR(16) UNIQUE,
	current_status      user_status,
 	api_token           VARCHAR(16),
 	firebase_id			    VARCHAR(255)
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
	inc_id              VARCHAR(12) PRIMARY KEY,
	requester_imei      VARCHAR(15),
	req_by_helper       BOOLEAN,
	time_start          TIMESTAMP WITHOUT TIME ZONE,
	time_end            TIMESTAMP WITHOUT TIME ZONE,
	is_resolved 		BOOLEAN
);

SELECT AddGeometryColumn('incidents', 'init_req_location', 4326, 'POINT', 2);

CREATE TABLE IF NOT EXISTS requests (
	req_id              serial PRIMARY KEY,
	u_id                INTEGER REFERENCES users(u_id),
	init_time           TIMESTAMP WITHOUT TIME ZONE,
	time_responded      TIMESTAMP WITHOUT TIME ZONE,
	response_val        BOOLEAN,
	has_kit		    BOOLEAN,
	inc_id              VARCHAR(12) REFERENCES incidents(inc_id)
);

--  Adds a 2 dimensional Geometry column of type point, in SRID 4326
SELECT AddGeometryColumn('requests', 'init_help_location', 4326, 'POINT', 2);


CREATE TABLE IF NOT EXISTS location (
	u_id                INTEGER UNIQUE REFERENCES users(u_id)
);

SELECT AddGeometryColumn('location', 'help_location', 4326, 'POINT', 2);


--	Add spatial index
CREATE INDEX ON location USING GIST(help_location);


--	Function to execute nearest neighbour query
--	parameter 1 is the requester location, parameter 2 is the search radius
CREATE OR REPLACE FUNCTION nearest_helpers(text, int)
RETURNS TABLE (
	candidate_id		INTEGER,
	distance 			INTEGER
)
AS $$
BEGIN
	RETURN query
	SELECT u_id, (ST_Distance(ST_Transform(help_location, 3005), ST_Transform(ST_GeomFromGeoJSON($1), 3005)):: int) AS distance
	FROM location NATURAL JOIN users
	WHERE ST_DWithin(ST_Transform(help_location, 3005), ST_Transform(ST_GeomFromGeoJSON($1), 3005), $2)
	AND current_status = 'active'
	ORDER BY ST_Distance(ST_Transform(help_location, 3005), ST_Transform(ST_GeomFromGeoJSON($1), 3005)) ASC;
END;
$$	LANGUAGE plpgsql;
