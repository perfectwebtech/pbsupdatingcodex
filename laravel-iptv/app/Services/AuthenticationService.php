<?php

namespace App\Services;

use App\Models\User;
use App\Models\LoginLog;
use App\Exceptions\AuthenticationException;
use Illuminate\Support\Facades\Hash;
use Tymon\JWTAuth\Facades\JWTAuth;

class AuthenticationService
{
    public function __construct(
        private GeoIPService $geoIPService,
        private LoggingService $loggingService
    ) {}

    /**
     * Authenticate user by username and password
     */
    public function authenticate(
        string $username,
        string $password,
        string $ipAddress,
        string $userAgent
    ): array {
        // Find user
        $user = User::where('username', $username)->first();

        // Log failed attempt
        if (!$user || !Hash::check($password, $user->password)) {
            $this->logLoginAttempt(null, $ipAddress, $userAgent, false, 'Invalid credentials');
            throw new AuthenticationException('Invalid username or password');
        }

        // Check if user is active
        if (!$user->is_active) {
            $this->logLoginAttempt($user->id, $ipAddress, $userAgent, false, 'User disabled');
            throw new AuthenticationException('User account is disabled');
        }

        if (!$user->admin_enabled) {
            $this->logLoginAttempt($user->id, $ipAddress, $userAgent, false, 'Admin disabled');
            throw new AuthenticationException('User account is suspended');
        }

        // Check expiration
        if ($user->isExpired()) {
            $this->logLoginAttempt($user->id, $ipAddress, $userAgent, false, 'Account expired');
            throw new AuthenticationException('User account has expired');
        }

        // Check IP whitelist
        if (!$user->isIpAllowed($ipAddress)) {
            $this->logLoginAttempt($user->id, $ipAddress, $userAgent, false, 'IP not allowed');
            throw new AuthenticationException('Access denied from this IP address');
        }

        // Check country restriction
        $countryCode = $this->geoIPService->getCountryCode($ipAddress);
        if (!$user->isCountryAllowed($countryCode)) {
            $this->logLoginAttempt($user->id, $ipAddress, $userAgent, false, 'Country not allowed');
            throw new AuthenticationException('Access denied from this country');
        }

        // Generate JWT token
        $token = JWTAuth::fromUser($user);

        // Log successful login
        $this->logLoginAttempt($user->id, $ipAddress, $userAgent, true);

        return [
            'access_token' => $token,
            'token_type' => 'bearer',
            'expires_in' => auth('api')->factory()->getTTL() * 60,
            'user' => [
                'id' => $user->id,
                'username' => $user->username,
                'max_connections' => $user->max_connections,
                'expires_at' => $user->expires_at?->toIso8601String(),
                'is_trial' => $user->is_trial,
            ],
        ];
    }

    /**
     * Validate token and get user
     */
    public function validateToken(string $token): User
    {
        try {
            $user = JWTAuth::setToken($token)->authenticate();

            if (!$user) {
                throw new AuthenticationException('Invalid token');
            }

            if (!$user->isValid()) {
                throw new AuthenticationException('User account is not valid');
            }

            return $user;
        } catch (\Exception $e) {
            throw new AuthenticationException('Token validation failed: ' . $e->getMessage());
        }
    }

    /**
     * Refresh token
     */
    public function refreshToken(string $token): array
    {
        try {
            $newToken = JWTAuth::setToken($token)->refresh();

            return [
                'access_token' => $newToken,
                'token_type' => 'bearer',
                'expires_in' => auth('api')->factory()->getTTL() * 60,
            ];
        } catch (\Exception $e) {
            throw new AuthenticationException('Token refresh failed');
        }
    }

    /**
     * Invalidate token (logout)
     */
    public function logout(string $token): void
    {
        try {
            JWTAuth::setToken($token)->invalidate();
        } catch (\Exception $e) {
            // Silent fail - token might already be invalid
        }
    }

    /**
     * Log login attempt
     */
    private function logLoginAttempt(
        ?int $userId,
        string $ipAddress,
        string $userAgent,
        bool $success,
        ?string $failureReason = null
    ): void {
        LoginLog::create([
            'user_id' => $userId,
            'ip_address' => $ipAddress,
            'user_agent' => $userAgent,
            'success' => $success,
            'failure_reason' => $failureReason,
        ]);
    }

    /**
     * Check if user can create new connection
     */
    public function canConnect(User $user): bool
    {
        return $user->canConnect();
    }

    /**
     * Get active connections for user
     */
    public function getActiveConnections(User $user): int
    {
        return $user->getActiveConnectionsCount();
    }
}
