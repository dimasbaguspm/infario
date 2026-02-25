CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    commit_hash VARCHAR(40) NOT NULL,
    commit_message TEXT,
    storage_path TEXT,
    public_url VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deployments_project_id ON deployments (project_id);
CREATE INDEX IF NOT EXISTS idx_deployments_public_url ON deployments (public_url);
CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON deployments (created_at DESC);
