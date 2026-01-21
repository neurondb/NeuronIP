-- Migration: Privacy Compliance Automation
-- Description: Adds DSAR, PIA, and consent management tables

-- DSAR Requests: Data Subject Access Requests
CREATE TABLE IF NOT EXISTS neuronip.dsar_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_type TEXT NOT NULL CHECK (request_type IN ('access', 'deletion', 'correction', 'portability')),
    subject_name TEXT NOT NULL,
    subject_email TEXT NOT NULL,
    subject_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'rejected')),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    request_details JSONB DEFAULT '{}',
    discovered_data JSONB DEFAULT '[]',
    response_data JSONB DEFAULT '{}',
    assigned_to TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.dsar_requests IS 'Data Subject Access Requests (GDPR/CCPA)';

CREATE INDEX IF NOT EXISTS idx_dsar_status ON neuronip.dsar_requests(status);
CREATE INDEX IF NOT EXISTS idx_dsar_subject_email ON neuronip.dsar_requests(subject_email);
CREATE INDEX IF NOT EXISTS idx_dsar_subject_id ON neuronip.dsar_requests(subject_id) WHERE subject_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_dsar_requested_at ON neuronip.dsar_requests(requested_at DESC);

-- PIA Requests: Privacy Impact Assessments
CREATE TABLE IF NOT EXISTS neuronip.pia_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    project_name TEXT NOT NULL,
    data_types JSONB DEFAULT '[]',
    data_subjects JSONB DEFAULT '[]',
    processing_purposes JSONB DEFAULT '[]',
    risk_level TEXT NOT NULL DEFAULT 'low' CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'review', 'approved', 'rejected')),
    submitted_by TEXT NOT NULL,
    reviewed_by TEXT,
    approved_by TEXT,
    assessment_results JSONB DEFAULT '{}',
    recommendations JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    reviewed_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.pia_requests IS 'Privacy Impact Assessment requests';

CREATE INDEX IF NOT EXISTS idx_pia_status ON neuronip.pia_requests(status);
CREATE INDEX IF NOT EXISTS idx_pia_risk_level ON neuronip.pia_requests(risk_level);
CREATE INDEX IF NOT EXISTS idx_pia_submitted_by ON neuronip.pia_requests(submitted_by);

-- Consent Records: Consent management
CREATE TABLE IF NOT EXISTS neuronip.consent_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id TEXT NOT NULL,
    subject_email TEXT NOT NULL,
    consent_type TEXT NOT NULL CHECK (consent_type IN ('marketing', 'analytics', 'data_sharing', 'processing')),
    purpose TEXT NOT NULL,
    consented BOOLEAN NOT NULL DEFAULT false,
    consented_at TIMESTAMPTZ,
    withdrawn_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    consent_method TEXT NOT NULL DEFAULT 'explicit' CHECK (consent_method IN ('explicit', 'implicit', 'opt_in', 'opt_out')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.consent_records IS 'Consent records for GDPR/CCPA compliance';

CREATE INDEX IF NOT EXISTS idx_consent_subject_id ON neuronip.consent_records(subject_id);
CREATE INDEX IF NOT EXISTS idx_consent_subject_email ON neuronip.consent_records(subject_email);
CREATE INDEX IF NOT EXISTS idx_consent_type ON neuronip.consent_records(consent_type);
CREATE INDEX IF NOT EXISTS idx_consent_consented ON neuronip.consent_records(consent_type, purpose, consented) WHERE consented = true;
CREATE INDEX IF NOT EXISTS idx_consent_expires_at ON neuronip.consent_records(expires_at) WHERE expires_at IS NOT NULL;
