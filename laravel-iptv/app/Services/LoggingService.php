<?php

namespace App\Services;

use App\Models\ClientLog;
use App\Models\UserSession;
use Illuminate\Support\Facades\Log;

class LoggingService
{
    /**
     * Log client action
     */
    public function logClientAction(
        ?int $userId,
        ?int $streamId,
        string $action,
        string $ipAddress,
        ?string $userAgent = null,
        ?array $extraData = null
    ): void {
        ClientLog::create([
            'user_id' => $userId,
            'stream_id' => $streamId,
            'ip_address' => $ipAddress,
            'user_agent' => $userAgent,
            'action' => $action,
            'query_string' => request()->getQueryString(),
            'extra_data' => $extraData,
        ]);
    }

    /**
     * Log stream start
     */
    public function logStreamStart(UserSession $session): void
    {
        $this->logClientAction(
            $session->user_id,
            $session->stream_id,
            'STREAM_START',
            $session->ip_address,
            $session->user_agent,
            [
                'container' => $session->container,
                'server_id' => $session->server_id,
            ]
        );

        Log::info("Stream started", [
            'user_id' => $session->user_id,
            'stream_id' => $session->stream_id,
            'ip' => $session->ip_address,
            'container' => $session->container,
        ]);
    }

    /**
     * Log stream end
     */
    public function logStreamEnd(UserSession $session): void
    {
        $duration = $session->getDuration();

        $this->logClientAction(
            $session->user_id,
            $session->stream_id,
            'STREAM_END',
            $session->ip_address,
            $session->user_agent,
            [
                'duration_seconds' => $duration,
                'bytes_transferred' => $session->bytes_transferred,
            ]
        );

        Log::info("Stream ended", [
            'user_id' => $session->user_id,
            'stream_id' => $session->stream_id,
            'duration' => $duration,
            'bytes' => $session->bytes_transferred,
        ]);
    }

    /**
     * Log authentication failure
     */
    public function logAuthFailure(
        string $username,
        string $ipAddress,
        string $reason
    ): void {
        $this->logClientAction(
            null,
            null,
            'AUTH_FAILED',
            $ipAddress,
            request()->userAgent(),
            [
                'username' => $username,
                'reason' => $reason,
            ]
        );

        Log::warning("Authentication failed", [
            'username' => $username,
            'ip' => $ipAddress,
            'reason' => $reason,
        ]);
    }

    /**
     * Log connection limit reached
     */
    public function logConnectionLimit(
        int $userId,
        string $ipAddress
    ): void {
        $this->logClientAction(
            $userId,
            null,
            'CONNECTION_LIMIT_REACHED',
            $ipAddress,
            request()->userAgent()
        );
    }

    /**
     * Log security event
     */
    public function logSecurityEvent(
        string $event,
        string $ipAddress,
        ?array $data = null
    ): void {
        $this->logClientAction(
            null,
            null,
            "SECURITY_{$event}",
            $ipAddress,
            request()->userAgent(),
            $data
        );

        Log::warning("Security event: {$event}", [
            'ip' => $ipAddress,
            'data' => $data,
        ]);
    }
}
