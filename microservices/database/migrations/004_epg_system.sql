-- Migration: EPG (Electronic Program Guide) System
-- Version: 004
-- Description: Add EPG management tables for TV program scheduling

-- =====================================================
-- EPG TABLES
-- =====================================================

-- EPG Programs (Electronic Program Guide)
CREATE TABLE IF NOT EXISTS epg_programs (\n    id BIGSERIAL PRIMARY KEY,
    stream_id BIGINT NOT NULL REFERENCES streams(id) ON DELETE CASCADE,

    -- Program details
    title VARCHAR(500) NOT NULL,
    description TEXT,
    category VARCHAR(100),

    -- Schedule
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration INT GENERATED ALWAYS AS (EXTRACT(EPOCH FROM (end_time - start_time))) STORED,

    -- Media
    image_url VARCHAR(500),
    icon_url VARCHAR(100),

    -- Metadata
    episode_number INT,
    season_number INT,
    year INT,
    rating VARCHAR(20), -- 'G', 'PG', 'PG-13', 'R', etc.
    directors JSONB, -- Array of director names
    actors JSONB, -- Array of actor names
    country VARCHAR(100),
    language VARCHAR(50),

    -- Additional info
    is_live BOOLEAN DEFAULT FALSE,
    is_repeat BOOLEAN DEFAULT FALSE,
    is_premiere BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Constraints
    CHECK (end_time > start_time),
    CHECK (duration >= 0)
);

-- EPG Import Sources (track where EPG data comes from)
CREATE TABLE IF NOT EXISTS epg_sources (
    id BIGSERIAL PRIMARY KEY,

    -- Source info
    name VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) NOT NULL, -- 'xmltv', 'json', 'api', 'manual'
    url VARCHAR(500),

    -- Configuration
    update_interval INT DEFAULT 3600, -- seconds between updates
    format_config JSONB, -- Custom format configuration

    -- Credentials (encrypted)
    credentials JSONB,

    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    last_sync TIMESTAMP,
    last_sync_status VARCHAR(50), -- 'success', 'failed', 'in_progress'
    last_error TEXT,

    -- Stats
    programs_imported INT DEFAULT 0,
    programs_updated INT DEFAULT 0,
    programs_failed INT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- EPG Import Log (track import history)
CREATE TABLE IF NOT EXISTS epg_import_log (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGINT REFERENCES epg_sources(id) ON DELETE SET NULL,

    -- Import details
    import_type VARCHAR(50), -- 'full', 'incremental'
    status VARCHAR(50), -- 'success', 'failed', 'partial'

    -- Stats
    programs_processed INT DEFAULT 0,
    programs_created INT DEFAULT 0,
    programs_updated INT DEFAULT 0,
    programs_deleted INT DEFAULT 0,
    programs_failed INT DEFAULT 0,

    -- Timing
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    duration_seconds INT,

    -- Error info
    error_message TEXT,
    error_details JSONB,

    -- Metadata
    import_file VARCHAR(500),
    imported_by BIGINT REFERENCES users(id) ON DELETE SET NULL,

    created_at TIMESTAMP DEFAULT NOW()
);

-- EPG Templates (reusable program templates)
CREATE TABLE IF NOT EXISTS epg_templates (
    id BIGSERIAL PRIMARY KEY,

    -- Template info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),

    -- Default values
    default_duration INT, -- in seconds
    default_rating VARCHAR(20),
    default_country VARCHAR(100),
    default_language VARCHAR(50),

    -- Template data
    template_data JSONB, -- Additional default fields

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- EPG User Reminders (users can set reminders for programs)
CREATE TABLE IF NOT EXISTS epg_reminders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    program_id BIGINT NOT NULL REFERENCES epg_programs(id) ON DELETE CASCADE,

    -- Reminder settings
    remind_minutes_before INT DEFAULT 10, -- Minutes before program starts
    reminder_sent BOOLEAN DEFAULT FALSE,
    reminder_sent_at TIMESTAMP,

    -- Notification preferences
    notify_email BOOLEAN DEFAULT FALSE,
    notify_push BOOLEAN DEFAULT TRUE,
    notify_sms BOOLEAN DEFAULT FALSE,

    -- Metadata
    notes TEXT,

    created_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, program_id)
);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- EPG Programs indexes
CREATE INDEX idx_epg_programs_stream ON epg_programs(stream_id);
CREATE INDEX idx_epg_programs_start_time ON epg_programs(start_time);
CREATE INDEX idx_epg_programs_end_time ON epg_programs(end_time);
CREATE INDEX idx_epg_programs_time_range ON epg_programs(stream_id, start_time, end_time);
CREATE INDEX idx_epg_programs_category ON epg_programs(category);
CREATE INDEX idx_epg_programs_title ON epg_programs USING gin(to_tsvector('english', title));

-- EPG Sources indexes
CREATE INDEX idx_epg_sources_active ON epg_sources(is_active);
CREATE INDEX idx_epg_sources_type ON epg_sources(source_type);

-- EPG Import Log indexes
CREATE INDEX idx_epg_import_log_source ON epg_import_log(source_id);
CREATE INDEX idx_epg_import_log_status ON epg_import_log(status);
CREATE INDEX idx_epg_import_log_started ON epg_import_log(started_at);

-- EPG Reminders indexes
CREATE INDEX idx_epg_reminders_user ON epg_reminders(user_id);
CREATE INDEX idx_epg_reminders_program ON epg_reminders(program_id);
CREATE INDEX idx_epg_reminders_pending ON epg_reminders(program_id, reminder_sent) WHERE reminder_sent = FALSE;

-- =====================================================
-- TRIGGERS FOR AUTO-UPDATE
-- =====================================================

-- Update epg_programs updated_at
CREATE TRIGGER epg_programs_updated_at
    BEFORE UPDATE ON epg_programs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Update epg_sources updated_at
CREATE TRIGGER epg_sources_updated_at
    BEFORE UPDATE ON epg_sources
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Calculate import duration on completion
CREATE OR REPLACE FUNCTION calculate_epg_import_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND OLD.completed_at IS NULL THEN
        NEW.duration_seconds = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at));
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER epg_import_duration
    BEFORE UPDATE ON epg_import_log
    FOR EACH ROW
    EXECUTE FUNCTION calculate_epg_import_duration();

-- =====================================================
-- VIEWS FOR COMMON QUERIES
-- =====================================================

-- Current and Upcoming Programs View
CREATE OR REPLACE VIEW epg_current_upcoming AS
SELECT
    p.*,
    s.name as stream_name,
    s.stream_url,
    CASE
        WHEN NOW() BETWEEN p.start_time AND p.end_time THEN 'now_playing'
        WHEN NOW() < p.start_time THEN 'upcoming'
        ELSE 'past'
    END as program_status,
    EXTRACT(EPOCH FROM (p.end_time - NOW())) as seconds_remaining
FROM epg_programs p
INNER JOIN streams s ON p.stream_id = s.id
WHERE p.end_time >= NOW() - INTERVAL '1 hour'
ORDER BY p.start_time;

-- EPG Schedule View (with stream info)
CREATE OR REPLACE VIEW epg_schedule AS
SELECT
    p.id,
    p.stream_id,
    s.name as stream_name,
    s.category_id,
    c.name as category_name,
    p.title,
    p.description,
    p.start_time,
    p.end_time,
    p.duration,
    p.category as program_category,
    p.image_url,
    p.rating,
    p.is_live,
    p.is_repeat,
    p.is_premiere
FROM epg_programs p
INNER JOIN streams s ON p.stream_id = s.id
LEFT JOIN categories c ON s.category_id = c.id
WHERE s.is_active = TRUE;

-- EPG Import Statistics View
CREATE OR REPLACE VIEW epg_import_stats AS
SELECT
    source_id,
    COUNT(*) as total_imports,
    COUNT(*) FILTER (WHERE status = 'success') as successful_imports,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_imports,
    SUM(programs_created) as total_programs_created,
    SUM(programs_updated) as total_programs_updated,
    SUM(programs_deleted) as total_programs_deleted,
    AVG(duration_seconds) as avg_duration_seconds,
    MAX(started_at) as last_import_time
FROM epg_import_log
GROUP BY source_id;

-- User Reminders with Program Info
CREATE OR REPLACE VIEW epg_user_reminders_view AS
SELECT
    r.id,
    r.user_id,
    u.username,
    u.email,
    r.program_id,
    p.title as program_title,
    p.start_time,
    p.end_time,
    s.name as stream_name,
    r.remind_minutes_before,
    r.reminder_sent,
    r.notify_email,
    r.notify_push,
    r.created_at,
    (p.start_time - (r.remind_minutes_before * INTERVAL '1 minute')) as reminder_time
FROM epg_reminders r
INNER JOIN users u ON r.user_id = u.id
INNER JOIN epg_programs p ON r.program_id = p.id
INNER JOIN streams s ON p.stream_id = s.id
WHERE p.end_time > NOW();

-- =====================================================
-- FUNCTIONS FOR EPG MANAGEMENT
-- =====================================================

-- Function to cleanup old EPG data
CREATE OR REPLACE FUNCTION cleanup_old_epg_data(days_to_keep INT DEFAULT 30)
RETURNS TABLE(deleted_programs BIGINT) AS $$
DECLARE
    deleted_count BIGINT;
BEGIN
    DELETE FROM epg_programs
    WHERE end_time < NOW() - (days_to_keep || ' days')::INTERVAL;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RETURN QUERY SELECT deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to get conflicting programs (overlapping time slots)
CREATE OR REPLACE FUNCTION get_epg_conflicts(p_stream_id BIGINT, p_start_time TIMESTAMP, p_end_time TIMESTAMP, p_exclude_id BIGINT DEFAULT NULL)
RETURNS TABLE(
    id BIGINT,
    title VARCHAR(500),
    start_time TIMESTAMP,
    end_time TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT p.id, p.title, p.start_time, p.end_time
    FROM epg_programs p
    WHERE p.stream_id = p_stream_id
    AND (p.id != p_exclude_id OR p_exclude_id IS NULL)
    AND (
        (p.start_time <= p_start_time AND p.end_time > p_start_time) OR
        (p.start_time < p_end_time AND p.end_time >= p_end_time) OR
        (p.start_time >= p_start_time AND p.end_time <= p_end_time)
    );
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE epg_programs IS 'Electronic Program Guide - TV program schedule';
COMMENT ON TABLE epg_sources IS 'EPG data import sources (XMLTV, APIs, etc.)';
COMMENT ON TABLE epg_import_log IS 'History of EPG data imports';
COMMENT ON TABLE epg_templates IS 'Reusable templates for EPG programs';
COMMENT ON TABLE epg_reminders IS 'User reminders for upcoming programs';

COMMENT ON COLUMN epg_programs.duration IS 'Auto-calculated program duration in seconds';
COMMENT ON FUNCTION cleanup_old_epg_data(INT) IS 'Removes EPG data older than specified days';
COMMENT ON FUNCTION get_epg_conflicts IS 'Finds overlapping programs for scheduling validation';
