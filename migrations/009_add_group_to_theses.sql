-- Добавляем group_id к таблиц theses для привязки курсовой к группе
ALTER TABLE theses ADD COLUMN IF NOT EXISTS group_id INT REFERENCES groups(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_theses_group ON theses(group_id);

-- Добавляем endpoint для самостоятельной записи студента в группу
-- Это будет handled в группе через POST /api/groups/{id}/join
