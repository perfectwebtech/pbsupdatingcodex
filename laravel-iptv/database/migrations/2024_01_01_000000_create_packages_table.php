<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('packages', function (Blueprint $table) {
            $table->id();
            $table->string('name');
            $table->text('description')->nullable();
            $table->decimal('price', 10, 2)->default(0);
            $table->integer('duration_days')->default(30);
            $table->integer('max_connections')->default(1);
            $table->boolean('is_trial')->default(false);
            $table->boolean('is_active')->default(true);
            $table->integer('trial_duration_days')->default(0);
            $table->json('allowed_output_formats')->nullable();
            $table->timestamps();

            $table->index('is_active');
        });

        // Pivot table for package streams
        Schema::create('package_stream', function (Blueprint $table) {
            $table->id();
            $table->foreignId('package_id')->constrained()->cascadeOnDelete();
            $table->foreignId('stream_id')->constrained()->cascadeOnDelete();
            $table->timestamps();

            $table->unique(['package_id', 'stream_id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('package_stream');
        Schema::dropIfExists('packages');
    }
};
