CREATE TABLE logs (
    log_id SERIAL PRIMARY KEY,
    command_number INTEGER,
    command_name VARCHAR(255),
    command TEXT,
    results TEXT,
    final_status VARCHAR(255),
    pipeline_fk_id INTEGER NOT NULL,
    FOREIGN KEY (pipeline_fk_id) REFERENCES pipelines(pipeline_id)
);