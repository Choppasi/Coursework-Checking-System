-- Seed initial admin user
INSERT INTO users (email, password_hash, role, full_name)
VALUES ('saltykovdanil373@gmail.com', '$2a$10$LeGVcx.MEdpBXeu02tZNDOM/ycpYoxUK4eMLVtUbPOAy.U/HOqHTW', 'admin', 'Салтыков Данил')
ON CONFLICT (email) DO NOTHING;
