<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Episode extends Model
{
    use HasFactory;

    protected $fillable = [
        'season_id',
        'stream_id',
        'episode_number',
        'title',
        'overview',
        'thumbnail_url',
        'runtime',
        'aired_at',
    ];

    protected $casts = [
        'aired_at' => 'datetime',
    ];

    // Relationships
    public function season()
    {
        return $this->belongsTo(Season::class);
    }

    public function stream()
    {
        return $this->belongsTo(Stream::class);
    }

    // Accessors
    public function getFullTitleAttribute(): string
    {
        return "S{$this->season->season_number}E{$this->episode_number} - {$this->title}";
    }
}
