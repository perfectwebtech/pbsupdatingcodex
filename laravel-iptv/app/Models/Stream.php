<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Stream extends Model
{
    use HasFactory;

    protected $fillable = [
        'name',
        'type',
        'category_id',
        'source_urls',
        'direct_source',
        'icon_url',
        'custom_sid',
        'order',
        'target_containers',
        'rtmp_output',
        'epg_id',
        'channel_id',
        'archive_server_id',
        'archive_duration_days',
        'metadata',
        'is_active',
        'added_at',
    ];

    protected $casts = [
        'source_urls' => 'array',
        'target_containers' => 'array',
        'metadata' => 'array',
        'direct_source' => 'boolean',
        'rtmp_output' => 'boolean',
        'is_active' => 'boolean',
        'added_at' => 'datetime',
    ];

    // Relationships
    public function category()
    {
        return $this->belongsTo(Category::class);
    }

    public function servers()
    {
        return $this->belongsToMany(Server::class, 'stream_server')
                    ->withPivot([
                        'pid',
                        'monitor_pid',
                        'status',
                        'on_demand',
                        'bitrate',
                        'delay_available_at',
                        'stream_info'
                    ])
                    ->withTimestamps();
    }

    public function archiveServer()
    {
        return $this->belongsTo(Server::class, 'archive_server_id');
    }

    public function packages()
    {
        return $this->belongsToMany(Package::class, 'package_stream')
                    ->withTimestamps();
    }

    public function sessions()
    {
        return $this->hasMany(UserSession::class);
    }

    public function activeSessions()
    {
        return $this->sessions()->whereNull('ended_at');
    }

    public function episodes()
    {
        return $this->hasMany(Episode::class);
    }

    // Scopes
    public function scopeActive($query)
    {
        return $query->where('is_active', true);
    }

    public function scopeOfType($query, string $type)
    {
        return $query->where('type', $type);
    }

    public function scopeLive($query)
    {
        return $query->where('type', 'live');
    }

    public function scopeVod($query)
    {
        return $query->where('type', 'vod');
    }

    public function scopeSeries($query)
    {
        return $query->where('type', 'series');
    }

    public function scopeRadio($query)
    {
        return $query->where('type', 'radio');
    }

    public function scopeInCategory($query, int $categoryId)
    {
        return $query->where('category_id', $categoryId);
    }

    // Methods
    public function isLive(): bool
    {
        return $this->type === 'live';
    }

    public function isVod(): bool
    {
        return $this->type === 'vod';
    }

    public function isSeries(): bool
    {
        return $this->type === 'series';
    }

    public function hasArchive(): bool
    {
        return $this->archive_duration_days > 0 && $this->archive_server_id !== null;
    }

    public function isOnlineOnServer(Server $server): bool
    {
        return $this->servers()
                    ->wherePivot('server_id', $server->id)
                    ->wherePivot('status', 'online')
                    ->exists();
    }

    public function getActiveViewersCount(): int
    {
        return $this->activeSessions()->count();
    }

    public function getStreamUrl(Server $server, string $container = 'ts'): ?string
    {
        if ($this->direct_source && !empty($this->source_urls)) {
            return $this->source_urls[0];
        }

        // Build streaming URL based on server and container
        $protocol = $server->protocol;
        $domain = $server->domain_name ?: $server->server_ip;
        $port = $protocol === 'https' ? $server->https_port : $server->http_port;

        return "{$protocol}://{$domain}:{$port}/stream/{$this->id}.{$container}";
    }
}
