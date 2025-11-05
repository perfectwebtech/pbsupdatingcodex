<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('epg_sources', function (Blueprint $table) {
            $table->id();
            $table->string('name');
            $table->string('url');
            $table->boolean('is_active')->default(true);
            $table->timestamp('last_updated_at')->nullable();
            $table->timestamps();
        });

        Schema::create('epg_data', function (Blueprint $table) {
            $table->id();
            $table->foreignId('epg_source_id')->constrained('epg_sources')->cascadeOnDelete();
            $table->string('channel_id');
            $table->string('title');
            $table->text('description')->nullable();
            $table->timestamp('start_time');
            $table->timestamp('end_time');
            $table->string('language', 10)->nullable();
            $table->json('extra_data')->nullable(); // genre, actors, etc.
            $table->timestamps();

            $table->index(['channel_id', 'start_time', 'end_time']);
            $table->index('start_time');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('epg_data');
        Schema::dropIfExists('epg_sources');
    }
};
