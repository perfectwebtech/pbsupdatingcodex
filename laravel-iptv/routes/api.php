<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\Api\AuthController;
use App\Http\Controllers\Api\StreamController;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
*/

// Public routes
Route::prefix('v1')->group(function () {
    // Authentication
    Route::post('/auth/login', [AuthController::class, 'login']);
    Route::post('/auth/refresh', [AuthController::class, 'refresh']);

    // Protected routes
    Route::middleware(['jwt.auth'])->group(function () {
        // Auth
        Route::post('/auth/logout', [AuthController::class, 'logout']);
        Route::get('/auth/me', [AuthController::class, 'me']);

        // Streams - with subscription check
        Route::middleware(['check.subscription'])->group(function () {
            // Categories
            Route::get('/categories', [StreamController::class, 'categories']);

            // Streams
            Route::get('/streams', [StreamController::class, 'index']);
            Route::get('/streams/{id}', [StreamController::class, 'show']);
            Route::get('/streams/{id}/url', [StreamController::class, 'getUrl']);

            // HLS Streaming
            Route::get('/stream/{id}/playlist.m3u8', [StreamController::class, 'playlist']);
            Route::get('/stream/{id}/segment/{segment}', [StreamController::class, 'segment']);
        });
    });
});

// Health check
Route::get('/health', function () {
    return response()->json([
        'status' => 'ok',
        'timestamp' => now()->toIso8601String(),
    ]);
});
