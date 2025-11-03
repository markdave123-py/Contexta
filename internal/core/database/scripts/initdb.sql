BEGIN;

-- Extensions
CREATE EXTENSION IF NOT EXISTS vector;      -- pgvector
CREATE EXTENSION IF NOT EXISTS pgcrypto;    -- gen_random_uuid()

-- Idempotent migration marker
CREATE TABLE IF NOT EXISTS contexta_meta (
  version        INT PRIMARY KEY,
  initialized_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO contexta_meta(version) VALUES (1) ON CONFLICT DO NOTHING;

-- Users
CREATE TABLE IF NOT EXISTS users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  first_name    TEXT NOT NULL DEFAULT '',
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Documents
CREATE TABLE IF NOT EXISTS documents (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  file_name    TEXT NOT NULL,
  storage_url  TEXT NOT NULL,
  source_type  TEXT NOT NULL CHECK (source_type IN ('upload','url')),
  status       TEXT NOT NULL CHECK (status IN ('uploaded','processing','ready','failed')),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_documents_user ON documents(user_id);

-- Document Chunks
CREATE TABLE IF NOT EXISTS document_chunks (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id  UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  position     INT  NOT NULL,
  text         TEXT NOT NULL,
  embedding    VECTOR(768) NOT NULL,
  token_count  INT  NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_chunks_docid           ON document_chunks(document_id);
CREATE INDEX IF NOT EXISTS idx_chunks_docid_position  ON document_chunks(document_id, position);

CREATE INDEX IF NOT EXISTS idx_chunks_embedding_ivf ON document_chunks
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Chat Sessions
CREATE TABLE IF NOT EXISTS chat_sessions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  document_id  UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_doc ON chat_sessions(user_id, document_id);

-- Chat Messages
CREATE TABLE IF NOT EXISTS chat_messages (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id  UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
  role        TEXT NOT NULL CHECK (role IN ('user','assistant')),
  content     TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id, created_at);

-- updated_at trigger for users and documents
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END; $$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_users_updated_at'
  ) THEN
    CREATE TRIGGER trg_users_updated_at
      BEFORE UPDATE ON users
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_documents_updated_at'
  ) THEN
    CREATE TRIGGER trg_documents_updated_at
      BEFORE UPDATE ON documents
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END $$;

COMMIT;
