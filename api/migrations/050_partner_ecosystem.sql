-- Migration: 050_partner_ecosystem.sql
-- Description: Partner ecosystem and marketplace

-- Partners table
CREATE TABLE IF NOT EXISTS neuronip.partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_name VARCHAR(255) NOT NULL UNIQUE,
    partner_type VARCHAR(50) NOT NULL, -- 'integration', 'reseller', 'technology', 'consulting'
    company_name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    website_url TEXT,
    logo_url TEXT,
    description TEXT,
    certification_level VARCHAR(50), -- 'certified', 'premium', 'enterprise'
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'active', 'suspended', 'inactive'
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_partners_type ON neuronip.partners(partner_type);
CREATE INDEX idx_partners_status ON neuronip.partners(status);
CREATE INDEX idx_partners_certification ON neuronip.partners(certification_level);

-- Partner integrations/extensions
CREATE TABLE IF NOT EXISTS neuronip.partner_extensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES neuronip.partners(id),
    extension_name VARCHAR(255) NOT NULL,
    extension_type VARCHAR(50) NOT NULL, -- 'connector', 'plugin', 'template', 'workflow'
    description TEXT,
    version VARCHAR(50) NOT NULL,
    extension_config JSONB NOT NULL,
    pricing_model VARCHAR(50), -- 'free', 'paid', 'subscription', 'usage_based'
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'published', 'deprecated'
    download_count INTEGER DEFAULT 0,
    rating_average DOUBLE PRECISION DEFAULT 0,
    rating_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_partner_extensions_partner ON neuronip.partner_extensions(partner_id);
CREATE INDEX idx_partner_extensions_type ON neuronip.partner_extensions(extension_type);
CREATE INDEX idx_partner_extensions_status ON neuronip.partner_extensions(status);

-- Marketplace listings
CREATE TABLE IF NOT EXISTS neuronip.marketplace_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    extension_id UUID NOT NULL REFERENCES neuronip.partner_extensions(id),
    listing_name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    tags TEXT[], -- Array of tags
    featured BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_marketplace_listings_extension ON neuronip.marketplace_listings(extension_id);
CREATE INDEX idx_marketplace_listings_category ON neuronip.marketplace_listings(category);
CREATE INDEX idx_marketplace_listings_featured ON neuronip.marketplace_listings(featured) WHERE featured = true;

COMMENT ON TABLE neuronip.partners IS 'Partner registry';
COMMENT ON TABLE neuronip.partner_extensions IS 'Partner-developed extensions';
COMMENT ON TABLE neuronip.marketplace_listings IS 'Marketplace listings for extensions';
