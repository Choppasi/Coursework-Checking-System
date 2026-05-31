z yCREATE TABLE IF NOT EXISTS point_results (
    id SERIAL PRIMARY KEY,
    point_id INT REFERENCES thesis_points(id) ON DELETE CASCADE,
    student_id INT REFERENCES users(id),
    content TEXT,
    file_url VARCHAR(500),
    file_name VARCHAR(255),
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    review TEXT,
    review_status VARCHAR(50) DEFAULT 'pending',
    reviewed_at TIMESTAMP,
    reviewed_by INT REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_point_results_point ON point_results(point_id);
CREATE INDEX IF NOT EXISTS idx_point_results_student ON point_results(student_id);
