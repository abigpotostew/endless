-- schema.sql
CREATE TABLE IF NOT EXISTS post (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS markov_chain_model (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert some sample data
-- INSERT INTO post (title, content) VALUES 
--     ('First Post', 'This is the content of the first post.'),
--     ('Second Post', 'This is the content of the second post.'),
--     ('Third Post', 'This is the content of the third post.');