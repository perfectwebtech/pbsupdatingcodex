<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        // Client activity logs
        Schema::create('client_logs', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->nullable()->constrained()->cascadeOnDelete();
            $table->foreignId('stream_id')->nullable()->constrained()->nullOnDelete();
            $table->ipAddress('ip_address');
            $table->text('user_agent')->nullable();
            $table->string('action', 50); // AUTH_FAILED, STREAM_START, etc.
            $table->text('query_string')->nullable();
            $table->json('extra_data')->nullable();
            $table->timestamp('created_at');

            $table->index(['user_id', 'created_at']);
            $table->index('action');
        });

        // Login logs
        Schema::create('login_logs', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->nullable()->constrained()->cascadeOnDelete();
            $table->ipAddress('ip_address');
            $table->text('user_agent')->nullable();
            $table->boolean('success')->default(false);
            $table->string('failure_reason')->nullable();
            $table->timestamp('created_at');

            $table->index(['user_id', 'created_at']);
            $table->index('ip_address');
        });

        // Blocked IPs
        Schema::create('blocked_ips', function (Blueprint $table) {
            $table->id();
            $table->ipAddress('ip_address')->unique();
            $table->string('reason');
            $table->text('notes')->nullable();
            $table->timestamp('blocked_at');
            $table->timestamp('expires_at')->nullable();
            $table->timestamps();

            $table->index('ip_address');
        });

        // Blocked User Agents
        Schema::create('blocked_user_agents', function (Blueprint $table) {
            $table->id();
            $table->string('user_agent');
            $table->boolean('exact_match')->default(false);
            $table->integer('blocked_attempts')->default(0);
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('blocked_user_agents');
        Schema::dropIfExists('blocked_ips');
        Schema::dropIfExists('login_logs');
        Schema::dropIfExists('client_logs');
    }
};
