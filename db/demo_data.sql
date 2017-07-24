--	DEMO DATA

--	USERS
INSERT INTO users(first_name, last_name, phone_number, current_status, api_token, firebase_id) VALUES
('library', 'uvic', '1234567890', 'active', '1234567890qwerty', 'null_id_1'),
('bookstore', 'uvic', '1234567891', 'active', '1234567890qwertu', 'null_id_2'),
('carsa', 'uvic', '1234567892', 'active', '1234567890qwerti', 'null_id_3'),
('felicitas', 'uvic', '1234567893', 'active', '1234567890qwerto', 'null_id_4'),
('bob wright', 'uvic', '1234567894', 'active', '1234567890qwertp', 'null_id_5'),
('maclauren', 'uvic', '1234567895', 'active', '1234567890qwerta', 'null_id_6'),
('finn. gardens', 'uvic', '1234567896', 'active', '1234567890qwerts', 'null_id_7'),
('dog park', 'uvic', '1234567897', 'active', '1234567890qwertd', 'null_id_8');

--	LOCATIONS
INSERT INTO location(u_id, help_location) VALUES
(1, '0101000020E6100000C712D6C6D8D35E405C1D0071573B4840'),	--	library
(2, '0101000020E6100000B2D47ABFD1D35EC0C5E57805A23B4840'),	--	bookstore
(3, '0101000020E6100000CFA2772AE0D35EC061191BBAD93B4840'),	--	carsa
(4, '0101000020E610000079EA9106B7D35EC0B16A10E6763B4840'),	--	fels
(5, '0101000020E6100000A6D590B8C7D35EC049D6E1E82A3B4840'),	--	bob wright
(6, '0101000020E610000016DD7A4D0FD45EC0FC5580EF363B4840'),	--	maclauren
(7, '0101000020E6100000A1D634EF38D45EC08ACC5CE0F23A4840'),	--	finnerty gardens
(8, '0101000020E6100000290989B48DD35EC07E5358A9A03A4840');	--	dog park
