<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use App\Services\GeoIPService;
use App\Services\LoggingService;

class GeoBlock
{
    public function __construct(
        private GeoIPService $geoIPService,
        private LoggingService $loggingService
    ) {}

    public function handle(Request $request, Closure $next)
    {
        $user = $request->user();

        if (!$user) {
            return $next($request);
        }

        $countryCode = $this->geoIPService->getCountryCode($request->ip());

        if ($countryCode && !$user->isCountryAllowed($countryCode)) {
            $this->loggingService->logSecurityEvent(
                'COUNTRY_BLOCKED',
                $request->ip(),
                [
                    'user_id' => $user->id,
                    'country_code' => $countryCode,
                ]
            );

            return response()->json([
                'success' => false,
                'message' => 'Access denied from this country',
            ], 403);
        }

        return $next($request);
    }
}
