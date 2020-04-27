CREATE TABLE users (
    id SERIAL NOT NULL ,
    selector varchar(255) NOT NULL UNIQUE,
    validator varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    email varchar(255) NOT NULL UNIQUE,
    postgres_user varchar(255) NOT NULL
);

CREATE TABLE sessions (
    id SERIAL NOT NULL,
    selector varchar(255) NOT NULL UNIQUE,
    validator varchar(255) NOT NULL,
    user_id int NOT NULL,
    exp TIMESTAMP
);

CREATE TABLE audit_query (
    id SERIAL NOT NULL,
    user_id int NOT NULL,
    postgres_user varchar(255) NOT NULL,
    status varchar(255),
    query varchar
);
