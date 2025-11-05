<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('streams', function (Blueprint $table) {
            $table->id();
            $table->string('name');
            $table->string('type'); // live, vod, series, radio
            $table->foreignId('category_id')->constrained()->cascadeOnDelete();

            // Source configuration
            $table->json('source_urls'); // Array of source URLs
            $table->boolean('direct_source')->default(false);

            // Display
            $table->string('icon_url')->nullable();
            $table->string('custom_sid')->nullable(); // For Enigma2
            $table->integer('order')->default(0);

            // Stream settings
            $table->json('target_containers')->nullable(); // ts, m3u8, rtmp
            $table->boolean('rtmp_output')->default(false);

            // EPG settings
            $table->string('epg_id')->nullable();
            $table->string('channel_id')->nullable();

            // Archive/DVR settings
            $table->foreignId('archive_server_id')->nullable()->constrained('servers')->nullOnDelete();
            $table->integer('archive_duration_days')->default(0);

            // Metadata (for VOD/Series)
            $table->json('metadata')->nullable(); // cast, plot, rating, etc.

            // Status
            $table->boolean('is_active')->default(true);
            $table->timestamp('added_at')->useCurrent();

            $table->timestamps();

            $table->index(['type', 'category_id', 'is_active']);
            $table->index('order');
            $table->index(['epg_id', 'channel_id']);
        });

        // Stream to Server mapping (which servers stream this)
        Schema::create('stream_server', function (Blueprint $table) {
            $table->id();
            $table->foreignId('stream_id')->constrained()->cascadeOnDelete();
            $table->foreignId('server_id')->constrained()->cascadeOnDelete();

            // Stream status on this server
            $table->integer('pid')->nullable();
            $table->integer('monitor_pid')->nullable();
            $table->enum('status', ['online', 'offline', 'error'])->default('offline');
            $table->boolean('on_demand')->default(false);
            $table->integer('bitrate')->nullable();
            $table->timestamp('delay_available_at')->nullable();
            $table->json('stream_info')->nullable();

            $table->timestamps();

            $table->unique(['stream_id', 'server_id']);
            $table->index(['stream_id', 'status']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('stream_server');
        Schema::dropIfExists('streams');
    }
};
