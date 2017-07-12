--	test data

--	test users
INSERT INTO users(first_name, last_name, phone_number, current_status, token)
VALUES('Tanner1', 'Zinck', '1234567891', 'active', '123456');

INSERT INTO users(first_name, last_name, phone_number, current_status, token)
VALUES('Tanner2', 'Zinck', '1234567892', 'active', '123456');

INSERT INTO users(first_name, last_name, phone_number, current_status, token)
VALUES('Tanner3', 'Zinck', '1234567893', 'active', '123456');

INSERT INTO users(first_name, last_name, phone_number, current_status, token)
VALUES('Tanner4', 'Zinck', '1234567894', 'active', '123456');

INSERT INTO users(first_name, last_name, phone_number, current_status, token)
VALUES('Tanner5', 'Zinck', '1234567895', 'active', '123456');

INSERT INTO users(first_name, last_name, phone_number, current_status, token)
VALUES('Tanner6', 'Zinck', '1234567896', 'active', '123456');


--	test locations
INSERT INTO location(u_id, help_location)
VALUES (
	1,
	ST_GeomFromGeoJSON(
		'{
			"type": "Point",
			"coordinates": [7.734375,51.835777520452],
			"crs":{"type":"name","properties":{"name":"EPSG:4326"}}
		}'
	)
),
(
	2,
	ST_GeomFromGeoJSON(
		'{
			"type": "Point",
			"coordinates": [7.834375,51.935777520452],
			"crs":{"type":"name","properties":{"name":"EPSG:4326"}}
		}'
	)
),
(
	3,
	ST_GeomFromGeoJSON(
		'{
			"type": "Point",
			"coordinates": [8.134375,52.835777520452],
			"crs":{"type":"name","properties":{"name":"EPSG:4326"}}
		}'
	)
),
(
	4,
	ST_GeomFromGeoJSON(
		'{
			"type": "Point",
			"coordinates": [7.735375,51.836777520452],
			"crs":{"type":"name","properties":{"name":"EPSG:4326"}}
		}'
	)
);

INSERT INTO location(u_id, help_location)
VALUES(
	5,
	ST_GeomFromGeoJSON(
		'{
			"type": "Point",
			"coordinates": [7.735375,51.836787520452],
			"crs":{"type":"name","properties":{"name":"EPSG:4326"}}
		}'
	)
);
