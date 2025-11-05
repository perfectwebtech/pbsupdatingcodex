<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('users', function (Blueprint $table) {
            $table->id();
            $table->string('username', 50)->unique();
            $table->string('password');
            $table->string('email')->nullable();

            // Subscription details
            $table->foreignId('package_id')->nullable()->constrained('packages')->nullOnDelete();
            $table->integer('max_connections')->default(1);
            $table->boolean('is_trial')->default(false);
            $table->timestamp('expires_at')->nullable();

            // Status flags
            $table->boolean('is_active')->default(true);
            $table->boolean('admin_enabled')->default(true);
            $table->boolean('is_restreamer')->default(false);

            // Device restrictions
            $table->boolean('is_mag')->default(false);
            $table->boolean('is_e2')->default(false);
            $table->boolean('is_stalker')->default(false);

            // IP and geo restrictions
            $table->json('allowed_ips')->nullable();
            $table->json('allowed_user_agents')->nullable();
            $table->ipAddress('forced_ip')->nullable();
            $table->string('forced_country', 2)->nullable();

            // ISP lock
            $table->boolean('is_isp_locked')->default(false);
            $table->string('isp_description')->nullable();

            // Server assignment
            $table->foreignId('forced_server_id')->nullable()->constrained('servers')->nullOnDelete();

            // Pairing for dual lines
            $table->foreignId('pair_id')->nullable()->constrained('users')->nullOnDelete();

            // User agent bypass
            $table->boolean('bypass_ua')->default(false);

            // Notes
            $table->text('admin_notes')->nullable();
            $table->text('reseller_notes')->nullable();

            // Created by
            $table->foreignId('created_by')->nullable()->constrained('users')->nullOnDelete();

            $table->timestamps();
            $table->softDeletes();

            // Indexes
            $table->index(['username', 'password']);
            $table->index('expires_at');
            $table->index(['is_active', 'admin_enabled']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('users');
    }
};
