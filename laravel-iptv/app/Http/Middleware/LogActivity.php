<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use App\Services\LoggingService;

class LogActivity
{
    public function __construct(
        private LoggingService $loggingService
    ) {}

    public function handle(Request $request, Closure $next)
    {
        $response = $next($request);

        // Log API access
        if ($user = $request->user()) {
            $this->loggingService->logClientAction(
                $user->id,
                $request->route('id'),
                'API_ACCESS',
                $request->ip(),
                $request->userAgent(),
                [
                    'method' => $request->method(),
                    'path' => $request->path(),
                    'status' => $response->status(),
                ]
            );
        }

        return $response;
    }
}
