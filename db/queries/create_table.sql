CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    movie_name VARCHAR(255) NOT NULL,
    synopsis TEXT NOT NULL,
    runtime_minutes INT NOT NULL,
    imdb_rating VARCHAR(20), --optional
    release_date DATE, --optional
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE
);