<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\SoftDeletes;
use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Notifications\Notifiable;
use Laravel\Sanctum\HasApiTokens;
use Tymon\JWTAuth\Contracts\JWTSubject;

class User extends Authenticatable implements JWTSubject
{
    use HasApiTokens, HasFactory, Notifiable, SoftDeletes;

    protected $fillable = [
        'username',
        'password',
        'email',
        'package_id',
        'max_connections',
        'is_trial',
        'expires_at',
        'is_active',
        'admin_enabled',
        'is_restreamer',
        'is_mag',
        'is_e2',
        'is_stalker',
        'allowed_ips',
        'allowed_user_agents',
        'forced_ip',
        'forced_country',
        'is_isp_locked',
        'isp_description',
        'forced_server_id',
        'pair_id',
        'bypass_ua',
        'admin_notes',
        'reseller_notes',
        'created_by',
    ];

    protected $hidden = [
        'password',
    ];

    protected $casts = [
        'allowed_ips' => 'array',
        'allowed_user_agents' => 'array',
        'expires_at' => 'datetime',
        'is_trial' => 'boolean',
        'is_active' => 'boolean',
        'admin_enabled' => 'boolean',
        'is_restreamer' => 'boolean',
        'is_mag' => 'boolean',
        'is_e2' => 'boolean',
        'is_stalker' => 'boolean',
        'is_isp_locked' => 'boolean',
        'bypass_ua' => 'boolean',
    ];

    // JWT methods
    public function getJWTIdentifier()
    {
        return $this->getKey();
    }

    public function getJWTCustomClaims()
    {
        return [
            'username' => $this->username,
            'is_restreamer' => $this->is_restreamer,
            'max_connections' => $this->max_connections,
        ];
    }

    // Relationships
    public function package()
    {
        return $this->belongsTo(Package::class);
    }

    public function sessions()
    {
        return $this->hasMany(UserSession::class);
    }

    public function activeSessions()
    {
        return $this->sessions()->whereNull('ended_at');
    }

    public function pairedUser()
    {
        return $this->belongsTo(User::class, 'pair_id');
    }

    public function forcedServer()
    {
        return $this->belongsTo(Server::class, 'forced_server_id');
    }

    public function creator()
    {
        return $this->belongsTo(User::class, 'created_by');
    }

    public function loginLogs()
    {
        return $this->hasMany(LoginLog::class);
    }

    public function clientLogs()
    {
        return $this->hasMany(ClientLog::class);
    }

    // Scopes
    public function scopeActive($query)
    {
        return $query->where('is_active', true)
                     ->where('admin_enabled', true);
    }

    public function scopeNotExpired($query)
    {
        return $query->where(function($q) {
            $q->whereNull('expires_at')
              ->orWhere('expires_at', '>', now());
        });
    }

    public function scopeExpired($query)
    {
        return $query->whereNotNull('expires_at')
                     ->where('expires_at', '<=', now());
    }

    public function scopeTrial($query)
    {
        return $query->where('is_trial', true);
    }

    // Business logic methods
    public function isExpired(): bool
    {
        return $this->expires_at && $this->expires_at->isPast();
    }

    public function isValid(): bool
    {
        return $this->is_active
            && $this->admin_enabled
            && !$this->isExpired();
    }

    public function canConnect(): bool
    {
        if (!$this->isValid()) {
            return false;
        }

        $activeCount = $this->activeSessions()->count();
        return $activeCount < $this->max_connections;
    }

    public function getActiveConnectionsCount(): int
    {
        return $this->activeSessions()->count();
    }

    public function hasAccessToStream(Stream $stream): bool
    {
        if (!$this->package) {
            return false;
        }

        return $this->package->streams()->where('streams.id', $stream->id)->exists();
    }

    public function isIpAllowed(string $ip): bool
    {
        if (empty($this->allowed_ips)) {
            return true;
        }

        return in_array($ip, $this->allowed_ips);
    }

    public function isCountryAllowed(string $countryCode): bool
    {
        if (!$this->forced_country) {
            return true;
        }

        return $this->forced_country === 'ALL' || $this->forced_country === $countryCode;
    }
}
