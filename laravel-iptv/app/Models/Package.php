<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Package extends Model
{
    use HasFactory;

    protected $fillable = [
        'name',
        'description',
        'price',
        'duration_days',
        'max_connections',
        'is_trial',
        'is_active',
        'trial_duration_days',
        'allowed_output_formats',
    ];

    protected $casts = [
        'price' => 'decimal:2',
        'is_trial' => 'boolean',
        'is_active' => 'boolean',
        'allowed_output_formats' => 'array',
    ];

    // Relationships
    public function users()
    {
        return $this->hasMany(User::class);
    }

    public function streams()
    {
        return $this->belongsToMany(Stream::class, 'package_stream')
                    ->withTimestamps();
    }

    // Scopes
    public function scopeActive($query)
    {
        return $query->where('is_active', true);
    }

    public function scopeTrial($query)
    {
        return $query->where('is_trial', true);
    }

    // Methods
    public function hasStream(Stream $stream): bool
    {
        return $this->streams()->where('streams.id', $stream->id)->exists();
    }

    public function addStream(Stream $stream): void
    {
        if (!$this->hasStream($stream)) {
            $this->streams()->attach($stream);
        }
    }

    public function removeStream(Stream $stream): void
    {
        $this->streams()->detach($stream);
    }

    public function getStreamsByType(string $type)
    {
        return $this->streams()->where('type', $type)->get();
    }
}
