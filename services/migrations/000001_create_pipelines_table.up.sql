CREATE TABLE pipelines (
    pipeline_id SERIAL PRIMARY KEY,
    status VARCHAR(255),
    repository VARCHAR(255),
    branch VARCHAR(255),
    commit VARCHAR(255)
);