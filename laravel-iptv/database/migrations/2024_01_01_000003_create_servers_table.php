<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('servers', function (Blueprint $table) {
            $table->id();
            $table->string('name');
            $table->ipAddress('server_ip');
            $table->string('domain_name')->nullable();

            // Ports
            $table->integer('http_port')->default(8000);
            $table->integer('https_port')->default(8443);
            $table->integer('rtmp_port')->default(1935);

            // Protocol
            $table->enum('protocol', ['http', 'https'])->default('https');

            // Capacity
            $table->integer('max_clients')->default(1000);
            $table->integer('network_guaranteed_speed')->default(1000); // Mbps

            // Status
            $table->boolean('is_online')->default(true);
            $table->boolean('is_enabled')->default(true);
            $table->boolean('timeshift_only')->default(false);

            // Hardware info
            $table->json('hardware_info')->nullable();

            // Network interface
            $table->string('network_interface')->default('eth0');

            // IP whitelist
            $table->json('whitelist_ips')->nullable();

            // GeoIP settings
            $table->boolean('enable_geoip')->default(false);
            $table->json('geoip_countries')->nullable();
            $table->enum('geoip_type', ['allow', 'deny', 'low_priority'])->default('allow');

            // ISP settings
            $table->boolean('enable_isp')->default(false);
            $table->json('isp_names')->nullable();
            $table->enum('isp_type', ['allow', 'deny', 'low_priority'])->default('allow');

            // Last check
            $table->timestamp('last_checked_at')->nullable();

            $table->timestamps();

            $table->index(['is_online', 'is_enabled']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('servers');
    }
};
