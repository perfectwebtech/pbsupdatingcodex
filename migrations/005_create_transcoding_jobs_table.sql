-- Migration: Create transcoding_jobs table
-- Description: Stores FFmpeg transcoding jobs for stream processing
-- Version: 005
-- Date: 2025-11-06

CREATE TABLE IF NOT EXISTS transcoding_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    -- Job Info
    job_name VARCHAR(255) NOT NULL,
    job_type ENUM('live_transcode', 'vod_transcode', 'recording', 'clip_generation') NOT NULL DEFAULT 'vod_transcode',
    status ENUM('pending', 'queued', 'processing', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending',
    priority INT NOT NULL DEFAULT 5 COMMENT '1=highest, 10=lowest',

    -- Source
    stream_id BIGINT UNSIGNED NULL,
    source_url TEXT NOT NULL,
    source_format VARCHAR(50) NULL,
    source_codec VARCHAR(50) NULL,
    source_bitrate INT NULL COMMENT 'in kbps',
    source_resolution VARCHAR(20) NULL COMMENT '1920x1080',
    source_duration INT NULL COMMENT 'in seconds',

    -- Target
    target_url TEXT NOT NULL,
    target_format VARCHAR(50) NOT NULL DEFAULT 'mp4',
    target_quality ENUM('sd', 'hd', 'fhd', 'uhd', 'custom') NOT NULL DEFAULT 'hd',
    target_preset VARCHAR(255) NULL COMMENT 'FFmpeg preset string',
    target_resolution VARCHAR(20) NULL COMMENT '1920x1080',
    target_bitrate INT NULL COMMENT 'in kbps',
    target_codec VARCHAR(50) NULL DEFAULT 'libx264',

    -- FFmpeg Settings
    ffmpeg_command TEXT NULL COMMENT 'Full FFmpeg command',
    hardware_acceleration BOOLEAN DEFAULT FALSE,
    encoder VARCHAR(50) DEFAULT 'libx264',
    audio_codec VARCHAR(50) DEFAULT 'aac',
    audio_bitrate INT DEFAULT 128 COMMENT 'in kbps',
    segment_duration INT DEFAULT 6 COMMENT 'for HLS, in seconds',
    extra_params TEXT NULL COMMENT 'Additional FFmpeg parameters',

    -- Progress Tracking
    progress DECIMAL(5,2) DEFAULT 0.00 COMMENT 'percentage 0.00-100.00',
    current_frame BIGINT DEFAULT 0,
    total_frames BIGINT DEFAULT 0,
    current_time INT DEFAULT 0 COMMENT 'in seconds',
    fps DECIMAL(8,2) DEFAULT 0.00,
    bitrate VARCHAR(20) NULL,
    speed DECIMAL(5,2) DEFAULT 0.00 COMMENT 'processing speed multiplier',

    -- Worker Info
    worker_id VARCHAR(100) NULL COMMENT 'ID of worker processing this job',
    worker_hostname VARCHAR(255) NULL,
    pid INT NULL COMMENT 'Process ID of FFmpeg',

    -- Timestamps
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    failed_at TIMESTAMP NULL,
    estimated_completion TIMESTAMP NULL,

    -- Error Handling
    error_message TEXT NULL,
    error_code VARCHAR(50) NULL,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,

    -- Output Info
    output_size BIGINT NULL COMMENT 'in bytes',
    output_duration INT NULL COMMENT 'in seconds',

    -- Metadata
    metadata JSON NULL COMMENT 'Additional metadata',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- Indexes
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_stream_id (stream_id),
    INDEX idx_job_type (job_type),
    INDEX idx_created_at (created_at),
    INDEX idx_worker_id (worker_id),
    INDEX idx_status_priority (status, priority),

    -- Foreign Keys
    FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create transcoding_queue view for pending/queued jobs
CREATE OR REPLACE VIEW transcoding_queue AS
SELECT
    id,
    job_name,
    job_type,
    status,
    priority,
    source_url,
    target_quality,
    progress,
    created_at,
    estimated_completion
FROM transcoding_jobs
WHERE status IN ('pending', 'queued', 'processing')
ORDER BY priority ASC, created_at ASC;

-- Create transcoding_stats table for analytics
CREATE TABLE IF NOT EXISTS transcoding_stats (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,

    -- Job Statistics
    total_jobs INT DEFAULT 0,
    completed_jobs INT DEFAULT 0,
    failed_jobs INT DEFAULT 0,
    cancelled_jobs INT DEFAULT 0,

    -- Performance
    avg_processing_time INT DEFAULT 0 COMMENT 'in seconds',
    total_processing_time BIGINT DEFAULT 0 COMMENT 'in seconds',
    avg_speed DECIMAL(5,2) DEFAULT 0.00,

    -- Resource Usage
    total_input_size BIGINT DEFAULT 0 COMMENT 'in bytes',
    total_output_size BIGINT DEFAULT 0 COMMENT 'in bytes',
    compression_ratio DECIMAL(5,2) DEFAULT 0.00,

    -- By Quality
    sd_jobs INT DEFAULT 0,
    hd_jobs INT DEFAULT 0,
    fhd_jobs INT DEFAULT 0,
    uhd_jobs INT DEFAULT 0,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_date (date),
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create trigger to update transcoding_stats
DELIMITER //

CREATE TRIGGER after_transcoding_job_complete
AFTER UPDATE ON transcoding_jobs
FOR EACH ROW
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        INSERT INTO transcoding_stats (
            date,
            total_jobs,
            completed_jobs,
            total_processing_time,
            total_input_size,
            total_output_size
        ) VALUES (
            CURDATE(),
            1,
            1,
            TIMESTAMPDIFF(SECOND, NEW.started_at, NEW.completed_at),
            0, -- Will be updated separately
            NEW.output_size
        )
        ON DUPLICATE KEY UPDATE
            total_jobs = total_jobs + 1,
            completed_jobs = completed_jobs + 1,
            total_processing_time = total_processing_time + TIMESTAMPDIFF(SECOND, NEW.started_at, NEW.completed_at),
            total_output_size = total_output_size + NEW.output_size,
            updated_at = CURRENT_TIMESTAMP;

        -- Update quality-specific counters
        IF NEW.target_quality = 'sd' THEN
            UPDATE transcoding_stats SET sd_jobs = sd_jobs + 1 WHERE date = CURDATE();
        ELSEIF NEW.target_quality = 'hd' THEN
            UPDATE transcoding_stats SET hd_jobs = hd_jobs + 1 WHERE date = CURDATE();
        ELSEIF NEW.target_quality = 'fhd' THEN
            UPDATE transcoding_stats SET fhd_jobs = fhd_jobs + 1 WHERE date = CURDATE();
        ELSEIF NEW.target_quality = 'uhd' THEN
            UPDATE transcoding_stats SET uhd_jobs = uhd_jobs + 1 WHERE date = CURDATE();
        END IF;
    END IF;

    IF NEW.status = 'failed' AND OLD.status != 'failed' THEN
        INSERT INTO transcoding_stats (date, total_jobs, failed_jobs)
        VALUES (CURDATE(), 1, 1)
        ON DUPLICATE KEY UPDATE
            total_jobs = total_jobs + 1,
            failed_jobs = failed_jobs + 1,
            updated_at = CURRENT_TIMESTAMP;
    END IF;
END//

DELIMITER ;

-- Insert sample transcoding jobs for testing
INSERT INTO transcoding_jobs (
    job_name,
    job_type,
    status,
    priority,
    stream_id,
    source_url,
    target_url,
    target_quality,
    target_resolution,
    target_bitrate,
    progress,
    created_at
) VALUES
('Movie_1080p_Transcode', 'vod_transcode', 'completed', 3, 1, 'https://cdn.example.com/source/movie1.mkv', 'https://cdn.example.com/output/movie1_1080p.mp4', 'fhd', '1920x1080', 5000, 100.00, NOW() - INTERVAL 2 HOUR),
('Series_S01E01_HD', 'vod_transcode', 'completed', 5, 2, 'https://cdn.example.com/source/series_s01e01.avi', 'https://cdn.example.com/output/series_s01e01_720p.mp4', 'hd', '1280x720', 2500, 100.00, NOW() - INTERVAL 1 HOUR),
('Live_Sports_Stream', 'live_transcode', 'processing', 1, 5, 'rtmp://source.example.com/live/sports', 'https://cdn.example.com/live/sports/index.m3u8', 'fhd', '1920x1080', 6000, 45.50, NOW() - INTERVAL 30 MINUTE),
('Documentary_4K', 'vod_transcode', 'queued', 7, NULL, 'https://cdn.example.com/source/documentary_4k.mov', 'https://cdn.example.com/output/documentary_4k.mp4', 'uhd', '3840x2160', 15000, 0.00, NOW() - INTERVAL 15 MINUTE),
('Clip_Highlights', 'clip_generation', 'pending', 4, 3, 'https://cdn.example.com/source/full_match.mp4', 'https://cdn.example.com/clips/highlights.mp4', 'hd', '1280x720', 3000, 0.00, NOW() - INTERVAL 5 MINUTE),
('Recording_CH101', 'recording', 'failed', 5, 8, 'rtmp://source.example.com/live/ch101', 'https://cdn.example.com/recordings/ch101_20251106.mp4', 'hd', '1280x720', 2500, 12.30, NOW() - INTERVAL 45 MINUTE);

-- Create indexes for better performance
CREATE INDEX idx_transcoding_jobs_composite ON transcoding_jobs(status, priority, created_at);
CREATE INDEX idx_transcoding_jobs_worker ON transcoding_jobs(worker_id, status);
