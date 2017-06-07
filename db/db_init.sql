CREATE TYPE user_status AS ENUM ('active', 'inactive', 'deleted');

CREATE TABLE users (
    u_id                serial PRIMARY KEY,
    first_name          VARHCAR(20),
    last_name           VARCHAR(20),
    phone_number        VARVHAR(16),
    current_status      user_status          
);

CREATE TABLE requests (
    req_id              serial PRIMARY KEY,
    u_id                INTEGER REFERENCES users(u_id)
);

CREATE TABLE incidents (
    requester_imei      VARCHAR(15),
    time_start          TIMESTAMP,
    time_end            TIMESTAMP
);

CREATE TABLE location (

);

