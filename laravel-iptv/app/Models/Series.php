<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

class Series extends Model
{
    use HasFactory;

    protected $fillable = [
        'title',
        'category_id',
        'plot',
        'cover_url',
        'backdrop_urls',
        'cast',
        'director',
        'genre',
        'release_date',
        'rating',
        'youtube_trailer',
        'episode_run_time',
        'is_active',
    ];

    protected $casts = [
        'backdrop_urls' => 'array',
        'rating' => 'decimal:1',
        'is_active' => 'boolean',
    ];

    // Relationships
    public function category()
    {
        return $this->belongsTo(Category::class);
    }

    public function seasons()
    {
        return $this->hasMany(Season::class);
    }

    // Scopes
    public function scopeActive($query)
    {
        return $query->where('is_active', true);
    }

    // Methods
    public function getTotalEpisodesCount(): int
    {
        return $this->seasons()->withCount('episodes')->get()->sum('episodes_count');
    }

    public function getTotalSeasonsCount(): int
    {
        return $this->seasons()->count();
    }
}
