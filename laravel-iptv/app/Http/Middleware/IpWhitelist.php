<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use App\Services\LoggingService;

class IpWhitelist
{
    public function __construct(
        private LoggingService $loggingService
    ) {}

    public function handle(Request $request, Closure $next)
    {
        $user = $request->user();

        if (!$user) {
            return $next($request);
        }

        if (!$user->isIpAllowed($request->ip())) {
            $this->loggingService->logSecurityEvent(
                'IP_NOT_ALLOWED',
                $request->ip(),
                ['user_id' => $user->id]
            );

            return response()->json([
                'success' => false,
                'message' => 'Access denied from this IP address',
            ], 403);
        }

        return $next($request);
    }
}
