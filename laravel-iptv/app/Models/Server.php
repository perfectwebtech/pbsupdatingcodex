<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Server extends Model
{
    use HasFactory;

    protected $fillable = [
        'name',
        'server_ip',
        'domain_name',
        'http_port',
        'https_port',
        'rtmp_port',
        'protocol',
        'max_clients',
        'network_guaranteed_speed',
        'is_online',
        'is_enabled',
        'timeshift_only',
        'hardware_info',
        'network_interface',
        'whitelist_ips',
        'enable_geoip',
        'geoip_countries',
        'geoip_type',
        'enable_isp',
        'isp_names',
        'isp_type',
        'last_checked_at',
    ];

    protected $casts = [
        'is_online' => 'boolean',
        'is_enabled' => 'boolean',
        'timeshift_only' => 'boolean',
        'hardware_info' => 'array',
        'whitelist_ips' => 'array',
        'enable_geoip' => 'boolean',
        'geoip_countries' => 'array',
        'enable_isp' => 'boolean',
        'isp_names' => 'array',
        'last_checked_at' => 'datetime',
    ];

    // Relationships
    public function streams()
    {
        return $this->belongsToMany(Stream::class, 'stream_server')
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

    public function sessions()
    {
        return $this->hasMany(UserSession::class);
    }

    public function activeSessions()
    {
        return $this->sessions()->whereNull('ended_at');
    }

    public function users()
    {
        return $this->hasMany(User::class, 'forced_server_id');
    }

    // Scopes
    public function scopeOnline($query)
    {
        return $query->where('is_online', true);
    }

    public function scopeEnabled($query)
    {
        return $query->where('is_enabled', true);
    }

    public function scopeAvailable($query)
    {
        return $query->where('is_online', true)
                     ->where('is_enabled', true)
                     ->where('timeshift_only', false);
    }

    // Methods
    public function isAvailable(): bool
    {
        return $this->is_online && $this->is_enabled && !$this->timeshift_only;
    }

    public function getActiveClientsCount(): int
    {
        return $this->activeSessions()->count();
    }

    public function hasCapacity(): bool
    {
        return $this->getActiveClientsCount() < $this->max_clients;
    }

    public function getLoadPercentage(): float
    {
        if ($this->max_clients == 0) {
            return 0;
        }

        return ($this->getActiveClientsCount() / $this->max_clients) * 100;
    }

    public function getSiteUrl(): string
    {
        $domain = $this->domain_name ?: $this->server_ip;
        $port = $this->protocol === 'https' ? $this->https_port : $this->http_port;

        return "{$this->protocol}://{$domain}:{$port}/";
    }

    public function isCountryAllowed(string $countryCode): bool
    {
        if (!$this->enable_geoip || empty($this->geoip_countries)) {
            return true;
        }

        $isInList = in_array($countryCode, $this->geoip_countries);

        return $this->geoip_type === 'allow' ? $isInList : !$isInList;
    }

    public function markOnline(): void
    {
        $this->update([
            'is_online' => true,
            'last_checked_at' => now(),
        ]);
    }

    public function markOffline(): void
    {
        $this->update([
            'is_online' => false,
            'last_checked_at' => now(),
        ]);
    }
}
