<?php

use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return response()->json([
        'app' => config('app.name'),
        'version' => '2.0.0',
        'api' => [
            'documentation' => url('/api/documentation'),
            'health' => url('/api/health'),
        ],
    ]);
});
