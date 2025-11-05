<?php

return [

    /*
    |--------------------------------------------------------------------------
    | FFmpeg Configuration
    |--------------------------------------------------------------------------
    */
    'ffmpeg' => [
        'path' => env('FFMPEG_PATH', '/usr/bin/ffmpeg'),
        'probe_path' => env('FFPROBE_PATH', '/usr/bin/ffprobe'),
        'timeout' => env('FFMPEG_TIMEOUT', 3600),
        'threads' => env('FFMPEG_THREADS', 4),
    ],

    /*
    |--------------------------------------------------------------------------
    | Paths Configuration
    |--------------------------------------------------------------------------
    */
    'paths' => [
        'streams' => env('STREAMS_PATH', storage_path('app/streams')),
        'archives' => env('ARCHIVES_PATH', storage_path('app/archives')),
        'temp' => env('TEMP_PATH', storage_path('app/temp')),
        'fonts' => env('FONTS_PATH', storage_path('app/fonts/free-sans.ttf')),
    ],

    /*
    |--------------------------------------------------------------------------
    | Stream Settings
    |--------------------------------------------------------------------------
    */
    'stream' => [
        'segment_duration' => env('STREAM_SEGMENT_DURATION', 10),
        'buffer_size' => env('STREAM_BUFFER_SIZE', 8192),
        'prebuffer_segments' => env('STREAM_PREBUFFER_SEGMENTS', 3),
        'playlist_size' => env('STREAM_PLAYLIST_SIZE', 10),
        'delete_threshold' => env('STREAM_DELETE_THRESHOLD', 3),
    ],

    /*
    |--------------------------------------------------------------------------
    | Transcoding Settings
    |--------------------------------------------------------------------------
    */
    'transcoding' => [
        'enabled' => env('ENABLE_TRANSCODING', true),
        'max_concurrent' => env('MAX_CONCURRENT_TRANSCODINGS', 5),
        'queue' => env('TRANSCODING_QUEUE', 'transcoding'),
        'profiles' => [
            'low' => [
                'video_bitrate' => '500k',
                'audio_bitrate' => '64k',
                'resolution' => '640x360',
                'fps' => 25,
            ],
            'medium' => [
                'video_bitrate' => '1500k',
                'audio_bitrate' => '128k',
                'resolution' => '1280x720',
                'fps' => 30,
            ],
            'high' => [
                'video_bitrate' => '3000k',
                'audio_bitrate' => '192k',
                'resolution' => '1920x1080',
                'fps' => 30,
            ],
            'ultra' => [
                'video_bitrate' => '6000k',
                'audio_bitrate' => '256k',
                'resolution' => '3840x2160',
                'fps' => 60,
            ],
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | DVR / Archive Settings
    |--------------------------------------------------------------------------
    */
    'dvr' => [
        'enabled' => env('ENABLE_DVR', true),
        'retention_days' => env('DVR_RETENTION_DAYS', 7),
        'storage_path' => env('DVR_STORAGE_PATH', storage_path('app/archives')),
    ],

    /*
    |--------------------------------------------------------------------------
    | CDN Configuration
    |--------------------------------------------------------------------------
    */
    'cdn' => [
        'enabled' => env('ENABLE_CDN', false),
        'provider' => env('CDN_PROVIDER', 's3'), // s3, cloudflare, akamai
        'url' => env('CDN_URL'),
        'signed_urls' => env('CDN_SIGNED_URLS', true),
        'url_expiry' => env('CDN_URL_EXPIRY', 7200), // 2 hours
    ],

    /*
    |--------------------------------------------------------------------------
    | Load Balancing
    |--------------------------------------------------------------------------
    */
    'load_balancing' => [
        'enabled' => env('ENABLE_LOAD_BALANCING', true),
        'strategy' => env('LOAD_BALANCE_STRATEGY', 'least_connections'), // least_connections, round_robin, weighted
        'health_check_interval' => env('HEALTH_CHECK_INTERVAL', 60),
    ],

    /*
    |--------------------------------------------------------------------------
    | Stream Types
    |--------------------------------------------------------------------------
    */
    'types' => [
        'live' => 'Live TV',
        'vod' => 'Video on Demand',
        'series' => 'TV Series',
        'radio' => 'Radio',
    ],

    /*
    |--------------------------------------------------------------------------
    | Supported Containers
    |--------------------------------------------------------------------------
    */
    'containers' => [
        'ts' => 'MPEG-TS',
        'm3u8' => 'HLS',
        'rtmp' => 'RTMP',
        'mp4' => 'MP4',
        'mkv' => 'Matroska',
    ],

    /*
    |--------------------------------------------------------------------------
    | GeoIP Configuration
    |--------------------------------------------------------------------------
    */
    'geoip' => [
        'database_path' => env('GEOIP_DATABASE_PATH', storage_path('app/GeoLite2.mmdb')),
        'cache_duration' => env('GEOIP_CACHE_DURATION', 86400), // 24 hours
    ],

    /*
    |--------------------------------------------------------------------------
    | Connection Limits
    |--------------------------------------------------------------------------
    */
    'limits' => [
        'max_connections_per_user' => env('MAX_CONNECTIONS_PER_USER', 3),
        'max_connections_per_ip' => env('MAX_CONNECTIONS_PER_IP', 5),
        'connection_timeout' => env('CONNECTION_TIMEOUT', 300), // 5 minutes
    ],

    /*
    |--------------------------------------------------------------------------
    | Analytics
    |--------------------------------------------------------------------------
    */
    'analytics' => [
        'enabled' => env('ENABLE_ANALYTICS', true),
        'driver' => env('ANALYTICS_DRIVER', 'clickhouse'), // database, clickhouse, redis
        'retention_days' => env('ANALYTICS_RETENTION_DAYS', 90),
    ],

];
