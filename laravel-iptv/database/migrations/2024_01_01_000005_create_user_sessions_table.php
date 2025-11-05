<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('user_sessions', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->foreignId('stream_id')->constrained()->cascadeOnDelete();
            $table->foreignId('server_id')->constrained()->cascadeOnDelete();

            // Connection details
            $table->ipAddress('ip_address');
            $table->text('user_agent')->nullable();
            $table->string('container', 20); // ts, m3u8, rtmp

            // Process ID
            $table->integer('pid')->nullable();

            // Geographic info
            $table->string('country_code', 2)->nullable();
            $table->string('isp')->nullable();

            // Device info
            $table->string('external_device')->nullable();

            // Session timing
            $table->timestamp('started_at');
            $table->timestamp('last_activity_at');
            $table->timestamp('ended_at')->nullable();

            // HLS specific
            $table->boolean('hls_end')->default(false);
            $table->timestamp('hls_last_read')->nullable();

            // Bandwidth tracking
            $table->bigInteger('bytes_transferred')->default(0);

            $table->timestamps();

            // Indexes
            $table->index(['user_id', 'ended_at']);
            $table->index(['stream_id', 'ended_at']);
            $table->index('started_at');
            $table->index(['user_id', 'stream_id', 'ended_at']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('user_sessions');
    }
};
