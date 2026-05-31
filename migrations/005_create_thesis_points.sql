CREATE TABLE IF NOT EXISTS thesis_points (
    id SERIAL PRIMARY KEY,
    thesis_id INT REFERENCES theses(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    "order" INT NOT NULL,
    deadline DATE,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_thesis_points_thesis ON thesis_points(thesis_id);
CREATE INDEX IF NOT EXISTS idx_thesis_points_order ON thesis_points("order");
