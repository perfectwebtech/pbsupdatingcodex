<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('series', function (Blueprint $table) {
            $table->id();
            $table->string('title');
            $table->foreignId('category_id')->constrained()->cascadeOnDelete();
            $table->text('plot')->nullable();
            $table->string('cover_url')->nullable();
            $table->json('backdrop_urls')->nullable();
            $table->string('cast')->nullable();
            $table->string('director')->nullable();
            $table->string('genre')->nullable();
            $table->string('release_date')->nullable();
            $table->decimal('rating', 3, 1)->default(0);
            $table->string('youtube_trailer')->nullable();
            $table->integer('episode_run_time')->nullable();
            $table->boolean('is_active')->default(true);
            $table->timestamps();

            $table->index(['category_id', 'is_active']);
        });

        Schema::create('seasons', function (Blueprint $table) {
            $table->id();
            $table->foreignId('series_id')->constrained()->cascadeOnDelete();
            $table->integer('season_number');
            $table->string('name')->nullable();
            $table->text('overview')->nullable();
            $table->string('cover_url')->nullable();
            $table->timestamps();

            $table->unique(['series_id', 'season_number']);
        });

        Schema::create('episodes', function (Blueprint $table) {
            $table->id();
            $table->foreignId('season_id')->constrained()->cascadeOnDelete();
            $table->foreignId('stream_id')->constrained()->cascadeOnDelete();
            $table->integer('episode_number');
            $table->string('title');
            $table->text('overview')->nullable();
            $table->string('thumbnail_url')->nullable();
            $table->integer('runtime')->nullable();
            $table->timestamp('aired_at')->nullable();
            $table->timestamps();

            $table->unique(['season_id', 'episode_number']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('episodes');
        Schema::dropIfExists('seasons');
        Schema::dropIfExists('series');
    }
};
