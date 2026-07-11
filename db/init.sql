-- Enable the pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE openai_embeddings (
    id SERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    embedding VECTOR(1536),
    metadata JSONB
);

CREATE TABLE ollama_embeddings (
    id SERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    embedding VECTOR(1024),
    metadata JSONB
);

CREATE INDEX ollama_embeddings_idx ON ollama_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

CREATE INDEX openai_embeddings_idx ON openai_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);