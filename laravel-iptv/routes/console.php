<?php

use Illuminate\Foundation\Inspiring;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Artisan::command('inspire', function () {
    $this->comment(Inspiring::quote());
})->purpose('Display an inspiring quote')->hourly();

// Scheduled tasks
Schedule::command('sessions:cleanup')->everyMinute();
Schedule::command('servers:check')->everyFiveMinutes();
Schedule::command('streams:monitor')->everyMinute();
