CREATE TABLE IF NOT EXISTS flag_dependencies (
    flag_id BIGINT NOT NULL,
    depends_on_id BIGINT NOT NULL,
    PRIMARY KEY (flag_id, depends_on_id),
    FOREIGN KEY (flag_id) REFERENCES flags(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_id) REFERENCES flags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_flag_dependencies_flag_id ON flag_dependencies(flag_id);
CREATE INDEX IF NOT EXISTS idx_flag_dependencies_depends_on_id ON flag_dependencies(depends_on_id); 