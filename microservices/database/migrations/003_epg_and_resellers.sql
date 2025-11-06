-- Migration: EPG and Reseller System
-- Version: 003
-- Description: Add EPG management and reseller hierarchy tables

-- =====================================================
-- RESELLER SYSTEM TABLES
-- =====================================================

-- Resellers table with hierarchical structure
CREATE TABLE IF NOT EXISTS resellers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES resellers(id) ON DELETE SET NULL,

    -- Credits and limits
    credits DECIMAL(10,2) DEFAULT 0 CHECK (credits >= 0),
    commission_rate DECIMAL(5,2) DEFAULT 0 CHECK (commission_rate >= 0 AND commission_rate <= 100),

    -- Permissions
    can_create_resellers BOOLEAN DEFAULT FALSE,
    max_users INT DEFAULT 100 CHECK (max_users >= 0),
    max_resellers INT DEFAULT 0 CHECK (max_resellers >= 0),

    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Constraints
    UNIQUE(user_id)
);

-- Reseller credits transaction log
CREATE TABLE IF NOT EXISTS reseller_credits_log (
    id BIGSERIAL PRIMARY KEY,
    reseller_id BIGINT NOT NULL REFERENCES resellers(id) ON DELETE CASCADE,

    -- Transaction details
    amount DECIMAL(10,2) NOT NULL,
    balance_before DECIMAL(10,2) NOT NULL,
    balance_after DECIMAL(10,2) NOT NULL,

    -- Transaction type: 'add', 'deduct', 'commission', 'refund'
    transaction_type VARCHAR(20) NOT NULL,
    description TEXT,

    -- Reference (invoice_id, payment_id, etc.)
    reference_id BIGINT,
    reference_type VARCHAR(50),

    -- Who made the change
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Reseller user assignments (track which users belong to which reseller)
CREATE TABLE IF NOT EXISTS reseller_assignments (
    id BIGSERIAL PRIMARY KEY,
    reseller_id BIGINT NOT NULL REFERENCES resellers(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Assignment details
    assigned_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,

    UNIQUE(user_id)
);

-- =====================================================
-- BILLING & INVOICING TABLES
-- =====================================================

-- Invoices
CREATE TABLE IF NOT EXISTS invoices (
    id BIGSERIAL PRIMARY KEY,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,

    -- Who
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reseller_id BIGINT REFERENCES resellers(id) ON DELETE SET NULL,

    -- What
    package_id BIGINT REFERENCES packages(id),
    description TEXT,

    -- Amounts
    subtotal DECIMAL(10,2) NOT NULL,
    tax DECIMAL(10,2) DEFAULT 0,
    discount DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) NOT NULL,

    -- Status: 'pending', 'paid', 'overdue', 'cancelled', 'refunded'
    status VARCHAR(20) DEFAULT 'pending',

    -- Dates
    issue_date TIMESTAMP DEFAULT NOW(),
    due_date TIMESTAMP,
    paid_at TIMESTAMP,

    -- Payment info
    payment_method VARCHAR(50),
    payment_gateway VARCHAR(50),
    transaction_id VARCHAR(255),

    -- Metadata
    metadata JSONB,
    notes TEXT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Payment methods (stored payment info)
CREATE TABLE IF NOT EXISTS payment_methods (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Type: 'credit_card', 'paypal', 'stripe', 'crypto', 'bank_transfer'
    type VARCHAR(50) NOT NULL,

    -- Encrypted payment details
    details JSONB, -- Card last 4 digits, PayPal email, etc.
    gateway_customer_id VARCHAR(255), -- Stripe customer ID, etc.

    -- Settings
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,

    -- Metadata
    nickname VARCHAR(100), -- e.g., "My Visa Card"

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Payment transactions
CREATE TABLE IF NOT EXISTS payment_transactions (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT REFERENCES invoices(id) ON DELETE SET NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payment_method_id BIGINT REFERENCES payment_methods(id) ON DELETE SET NULL,

    -- Transaction details
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',

    -- Status: 'pending', 'completed', 'failed', 'refunded'
    status VARCHAR(20) DEFAULT 'pending',

    -- Gateway info
    gateway VARCHAR(50), -- 'stripe', 'paypal', etc.
    gateway_transaction_id VARCHAR(255),
    gateway_response JSONB,

    -- Metadata
    ip_address VARCHAR(45),
    user_agent TEXT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- =====================================================
-- SERIES & EPISODES TABLES
-- =====================================================

-- TV Series
CREATE TABLE IF NOT EXISTS series (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id BIGINT REFERENCES categories(id),

    -- Media
    cover_url VARCHAR(500),
    backdrop_url VARCHAR(500),
    trailer_url VARCHAR(500),

    -- Metadata
    rating DECIMAL(3,2), -- e.g., 8.5
    release_year INT,
    genre VARCHAR(100),
    cast JSONB, -- Array of actors
    director VARCHAR(255),
    producer VARCHAR(255),

    -- Settings
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,

    -- Stats
    total_seasons INT DEFAULT 0,
    total_episodes INT DEFAULT 0,
    view_count BIGINT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Episodes
CREATE TABLE IF NOT EXISTS episodes (
    id BIGSERIAL PRIMARY KEY,
    series_id BIGINT NOT NULL REFERENCES series(id) ON DELETE CASCADE,

    -- Episode info
    season INT NOT NULL CHECK (season > 0),
    episode INT NOT NULL CHECK (episode > 0),
    title VARCHAR(255) NOT NULL,
    description TEXT,

    -- Media
    stream_url VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    duration INT, -- in seconds

    -- Metadata
    air_date DATE,
    rating DECIMAL(3,2),

    -- Settings
    is_active BOOLEAN DEFAULT TRUE,

    -- Stats
    view_count BIGINT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(series_id, season, episode)
);

-- =====================================================
-- DEVICE MANAGEMENT TABLES
-- =====================================================

-- User devices (MAG, Enigma2, Android, etc.)
CREATE TABLE IF NOT EXISTS devices (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Device identification
    device_type VARCHAR(50) NOT NULL, -- 'mag', 'enigma2', 'android', 'ios', 'web', 'smart_tv'
    device_id VARCHAR(100) NOT NULL, -- Unique device identifier
    mac_address VARCHAR(17),

    -- Device details
    model VARCHAR(100),
    manufacturer VARCHAR(100),
    os_version VARCHAR(50),
    app_version VARCHAR(50),

    -- Connection info
    last_ip VARCHAR(45),
    last_user_agent TEXT,
    last_seen TIMESTAMP,

    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_blocked BOOLEAN DEFAULT FALSE,

    -- Metadata
    nickname VARCHAR(100), -- User-friendly name
    notes TEXT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, device_id)
);

-- Device sessions (streaming sessions per device)
CREATE TABLE IF NOT EXISTS device_sessions (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stream_id BIGINT REFERENCES streams(id) ON DELETE SET NULL,

    -- Session details
    session_id VARCHAR(100) UNIQUE NOT NULL,
    ip_address VARCHAR(45),

    -- Timestamps
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP,
    last_activity TIMESTAMP DEFAULT NOW(),

    -- Stats
    bandwidth_used BIGINT DEFAULT 0, -- in bytes
    duration INT DEFAULT 0 -- in seconds
);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Resellers indexes
CREATE INDEX idx_resellers_user_id ON resellers(user_id);
CREATE INDEX idx_resellers_parent_id ON resellers(parent_id);
CREATE INDEX idx_resellers_active ON resellers(is_active);
CREATE INDEX idx_reseller_credits_log_reseller ON reseller_credits_log(reseller_id);
CREATE INDEX idx_reseller_credits_log_created ON reseller_credits_log(created_at);
CREATE INDEX idx_reseller_assignments_reseller ON reseller_assignments(reseller_id);

-- Billing indexes
CREATE INDEX idx_invoices_user_id ON invoices(user_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_created ON invoices(created_at);
CREATE INDEX idx_payment_methods_user ON payment_methods(user_id);
CREATE INDEX idx_payment_transactions_user ON payment_transactions(user_id);
CREATE INDEX idx_payment_transactions_invoice ON payment_transactions(invoice_id);

-- Series indexes
CREATE INDEX idx_series_category ON series(category_id);
CREATE INDEX idx_series_active ON series(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_series_featured ON series(is_featured) WHERE is_active = TRUE;
CREATE INDEX idx_episodes_series ON episodes(series_id);
CREATE INDEX idx_episodes_season ON episodes(series_id, season);

-- Device indexes
CREATE INDEX idx_devices_user ON devices(user_id);
CREATE INDEX idx_devices_type ON devices(device_type);
CREATE INDEX idx_devices_active ON devices(is_active);
CREATE INDEX idx_device_sessions_device ON device_sessions(device_id);
CREATE INDEX idx_device_sessions_user ON device_sessions(user_id);
CREATE INDEX idx_device_sessions_active ON device_sessions(ended_at) WHERE ended_at IS NULL;

-- =====================================================
-- TRIGGERS FOR AUTO-UPDATE
-- =====================================================

-- Update resellers updated_at
CREATE OR REPLACE FUNCTION update_resellers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER resellers_updated_at
    BEFORE UPDATE ON resellers
    FOR EACH ROW
    EXECUTE FUNCTION update_resellers_updated_at();

-- Update invoices updated_at
CREATE TRIGGER invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Update series updated_at
CREATE TRIGGER series_updated_at
    BEFORE UPDATE ON series
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Auto-increment series stats when episode is added
CREATE OR REPLACE FUNCTION update_series_episode_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE series
        SET total_episodes = total_episodes + 1,
            total_seasons = (
                SELECT MAX(season) FROM episodes WHERE series_id = NEW.series_id
            )
        WHERE id = NEW.series_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE series
        SET total_episodes = total_episodes - 1,
            total_seasons = (
                SELECT COALESCE(MAX(season), 0) FROM episodes WHERE series_id = OLD.series_id
            )
        WHERE id = OLD.series_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_series_stats
    AFTER INSERT OR DELETE ON episodes
    FOR EACH ROW
    EXECUTE FUNCTION update_series_episode_count();

-- =====================================================
-- VIEWS FOR COMMON QUERIES
-- =====================================================

-- Reseller hierarchy view
CREATE OR REPLACE VIEW reseller_hierarchy AS
WITH RECURSIVE reseller_tree AS (
    -- Base case: top-level resellers
    SELECT
        id,
        user_id,
        parent_id,
        credits,
        commission_rate,
        0 as level,
        ARRAY[id] as path
    FROM resellers
    WHERE parent_id IS NULL

    UNION ALL

    -- Recursive case: child resellers
    SELECT
        r.id,
        r.user_id,
        r.parent_id,
        r.credits,
        r.commission_rate,
        rt.level + 1,
        rt.path || r.id
    FROM resellers r
    INNER JOIN reseller_tree rt ON r.parent_id = rt.id
)
SELECT * FROM reseller_tree;

-- Active device sessions view
CREATE OR REPLACE VIEW active_device_sessions AS
SELECT
    ds.*,
    u.username,
    d.device_type,
    d.model,
    s.name as stream_name
FROM device_sessions ds
INNER JOIN users u ON ds.user_id = u.id
INNER JOIN devices d ON ds.device_id = d.id
LEFT JOIN streams s ON ds.stream_id = s.id
WHERE ds.ended_at IS NULL;

-- Invoice summary view
CREATE OR REPLACE VIEW invoice_summary AS
SELECT
    i.*,
    u.username,
    u.email,
    p.name as package_name,
    r.id as reseller_id
FROM invoices i
INNER JOIN users u ON i.user_id = u.id
LEFT JOIN packages p ON i.package_id = p.id
LEFT JOIN reseller_assignments ra ON ra.user_id = u.id
LEFT JOIN resellers r ON ra.reseller_id = r.id;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE resellers IS 'Reseller management with hierarchical structure';
COMMENT ON TABLE reseller_credits_log IS 'Transaction log for reseller credits';
COMMENT ON TABLE invoices IS 'Customer invoices and billing';
COMMENT ON TABLE series IS 'TV series metadata';
COMMENT ON TABLE episodes IS 'Individual episodes within series';
COMMENT ON TABLE devices IS 'User devices (MAG, Android, etc.)';
COMMENT ON TABLE device_sessions IS 'Active streaming sessions per device';
