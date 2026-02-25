-- Enable UUID extension if not already present
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Projects: The owner of everything
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    git_url VARCHAR(255) NOT NULL,
    git_provider VARCHAR(20) NOT NULL DEFAULT 'github', 
    primary_branch VARCHAR(50) DEFAULT 'main',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Deployments: The immutable snapshots of a project
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    commit_hash VARCHAR(40) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- 0:Queued, 1:Building, 2:Ready, 3:Failed
    preview_url VARCHAR(255) UNIQUE,
    storage_key TEXT, -- Path to files in S3/Local storage
    metadata_json JSONB, -- For extra git info like commit message
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


CREATE INDEX idx_deployments_preview_url ON deployments(preview_url) WHERE status = 2;
CREATE INDEX idx_deployments_project_latest ON deployments(project_id, created_at DESC);