<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class UserSession extends Model
{
    use HasFactory;

    protected $fillable = [
        'user_id',
        'stream_id',
        'server_id',
        'ip_address',
        'user_agent',
        'container',
        'pid',
        'country_code',
        'isp',
        'external_device',
        'started_at',
        'last_activity_at',
        'ended_at',
        'hls_end',
        'hls_last_read',
        'bytes_transferred',
    ];

    protected $casts = [
        'started_at' => 'datetime',
        'last_activity_at' => 'datetime',
        'ended_at' => 'datetime',
        'hls_last_read' => 'datetime',
        'hls_end' => 'boolean',
    ];

    // Relationships
    public function user()
    {
        return $this->belongsTo(User::class);
    }

    public function stream()
    {
        return $this->belongsTo(Stream::class);
    }

    public function server()
    {
        return $this->belongsTo(Server::class);
    }

    // Scopes
    public function scopeActive($query)
    {
        return $query->whereNull('ended_at');
    }

    public function scopeEnded($query)
    {
        return $query->whereNotNull('ended_at');
    }

    public function scopeForUser($query, User $user)
    {
        return $query->where('user_id', $user->id);
    }

    public function scopeForStream($query, Stream $stream)
    {
        return $query->where('stream_id', $stream->id);
    }

    public function scopeByContainer($query, string $container)
    {
        return $query->where('container', $container);
    }

    // Methods
    public function isActive(): bool
    {
        return $this->ended_at === null;
    }

    public function getDuration(): int
    {
        $endTime = $this->ended_at ?? now();
        return $endTime->diffInSeconds($this->started_at);
    }

    public function updateActivity(): void
    {
        $this->last_activity_at = now();
        $this->save();
    }

    public function end(): void
    {
        $this->ended_at = now();
        $this->save();
    }

    public function isHls(): bool
    {
        return $this->container === 'm3u8';
    }

    public function isRtmp(): bool
    {
        return $this->container === 'rtmp';
    }

    public function shouldBeTerminated(int $timeoutSeconds = 300): bool
    {
        if (!$this->isActive()) {
            return false;
        }

        return $this->last_activity_at->diffInSeconds(now()) > $timeoutSeconds;
    }
}
