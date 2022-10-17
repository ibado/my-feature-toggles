CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    email VARCHAR (50) UNIQUE NOT NULL,
    password_hash VARCHAR (250) NOT NULL
);

CREATE TABLE IF NOT EXISTS toggles (
    id VARCHAR (50) NOT NULL,
    value VARCHAR (250) NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
