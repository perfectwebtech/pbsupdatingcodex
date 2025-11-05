<?php

namespace App\Services;

use App\Models\User;
use App\Models\Stream;
use App\Models\Server;
use App\Models\UserSession;
use App\Exceptions\StreamingException;
use Illuminate\Support\Facades\Log;

class StreamingService
{
    public function __construct(
        private LoadBalancerService $loadBalancer,
        private FFmpegService $ffmpegService,
        private LoggingService $loggingService
    ) {}

    /**
     * Get stream for user
     */
    public function getStream(
        User $user,
        int $streamId,
        string $container,
        string $ipAddress,
        string $userAgent
    ): array {
        // Find stream
        $stream = Stream::active()->find($streamId);

        if (!$stream) {
            throw new StreamingException('Stream not found');
        }

        // Check user access
        if (!$user->hasAccessToStream($stream)) {
            $this->loggingService->logClientAction(
                $user->id,
                $streamId,
                'NOT_IN_PACKAGE',
                $ipAddress
            );
            throw new StreamingException('You do not have access to this stream');
        }

        // Check connection limit
        if (!$user->canConnect()) {
            $this->loggingService->logClientAction(
                $user->id,
                $streamId,
                'CONNECTION_LIMIT',
                $ipAddress
            );
            throw new StreamingException('Maximum connections reached');
        }

        // Select best server using load balancer
        $server = $this->loadBalancer->selectBestServer($stream, $user);

        if (!$server) {
            throw new StreamingException('No available servers');
        }

        // For direct source, return URL
        if ($stream->direct_source) {
            return [
                'type' => 'redirect',
                'url' => $stream->source_urls[0],
            ];
        }

        // Get stream URL
        $streamUrl = $this->getStreamUrl($stream, $server, $container);

        // Create session
        $session = $this->createSession(
            $user,
            $stream,
            $server,
            $ipAddress,
            $userAgent,
            $container
        );

        return [
            'type' => 'stream',
            'url' => $streamUrl,
            'session_id' => $session->id,
            'container' => $container,
        ];
    }

    /**
     * Generate HLS playlist
     */
    public function getHlsPlaylist(
        Stream $stream,
        Server $server,
        User $user
    ): string {
        $playlistPath = config('streaming.paths.streams') . "/{$stream->id}.m3u8";

        if (!file_exists($playlistPath)) {
            // Start stream if on-demand
            if ($this->shouldStartOnDemand($stream, $server)) {
                $this->ffmpegService->startStream($stream, $server);

                // Wait for playlist
                $attempts = 0;
                while (!file_exists($playlistPath) && $attempts < 20) {
                    usleep(500000); // 0.5 seconds
                    $attempts++;
                }
            }

            if (!file_exists($playlistPath)) {
                throw new StreamingException('Stream playlist not available');
            }
        }

        // Generate authenticated playlist
        return $this->generateAuthenticatedPlaylist(
            $playlistPath,
            $user,
            $stream
        );
    }

    /**
     * Get HLS segment
     */
    public function getHlsSegment(
        string $segmentName,
        User $user,
        Stream $stream
    ): array {
        $segmentPath = config('streaming.paths.streams') . "/{$segmentName}";

        if (!file_exists($segmentPath)) {
            throw new StreamingException('Segment not found');
        }

        // Verify segment belongs to this stream
        $streamId = (int)explode('_', $segmentName)[0];
        if ($streamId !== $stream->id) {
            throw new StreamingException('Invalid segment');
        }

        return [
            'path' => $segmentPath,
            'size' => filesize($segmentPath),
            'content_type' => 'video/mp2t',
        ];
    }

    /**
     * Create user session
     */
    private function createSession(
        User $user,
        Stream $stream,
        Server $server,
        string $ipAddress,
        string $userAgent,
        string $container
    ): UserSession {
        // Get geographic info
        $geoip = app(GeoIPService::class);
        $countryCode = $geoip->getCountryCode($ipAddress);
        $isp = $geoip->getISP($ipAddress);

        return UserSession::create([
            'user_id' => $user->id,
            'stream_id' => $stream->id,
            'server_id' => $server->id,
            'ip_address' => $ipAddress,
            'user_agent' => $userAgent,
            'container' => $container,
            'country_code' => $countryCode,
            'isp' => $isp,
            'started_at' => now(),
            'last_activity_at' => now(),
        ]);
    }

    /**
     * Close user session
     */
    public function closeSession(int $sessionId): void
    {
        $session = UserSession::find($sessionId);

        if ($session && $session->isActive()) {
            $session->end();

            // Log to analytics
            $this->loggingService->logStreamEnd($session);
        }
    }

    /**
     * Close oldest connection for user
     */
    public function closeOldestConnection(User $user): void
    {
        $oldestSession = $user->activeSessions()
            ->orderBy('started_at', 'asc')
            ->first();

        if ($oldestSession) {
            $this->closeSession($oldestSession->id);
        }
    }

    /**
     * Get stream URL
     */
    private function getStreamUrl(
        Stream $stream,
        Server $server,
        string $container
    ): string {
        return $stream->getStreamUrl($server, $container);
    }

    /**
     * Check if stream should start on demand
     */
    private function shouldStartOnDemand(Stream $stream, Server $server): bool
    {
        return $stream->servers()
            ->wherePivot('server_id', $server->id)
            ->wherePivot('on_demand', true)
            ->exists();
    }

    /**
     * Generate authenticated HLS playlist
     */
    private function generateAuthenticatedPlaylist(
        string $playlistPath,
        User $user,
        Stream $stream
    ): string {
        $content = file_get_contents($playlistPath);
        $lines = explode("\n", $content);
        $output = [];

        foreach ($lines as $line) {
            if (preg_match('/\.ts$/', $line)) {
                // Add authentication token to segment URL
                $token = md5($line . $user->username . config('app.key'));
                $output[] = "/stream/{$stream->id}/segment/{$line}?token={$token}";
            } else {
                $output[] = $line;
            }
        }

        return implode("\n", $output);
    }

    /**
     * Update session activity
     */
    public function updateSessionActivity(int $sessionId): void
    {
        $session = UserSession::find($sessionId);
        if ($session) {
            $session->updateActivity();
        }
    }

    /**
     * Terminate inactive sessions
     */
    public function terminateInactiveSessions(int $timeoutSeconds = 300): int
    {
        $sessions = UserSession::active()
            ->where('last_activity_at', '<', now()->subSeconds($timeoutSeconds))
            ->get();

        foreach ($sessions as $session) {
            $this->closeSession($session->id);
        }

        return $sessions->count();
    }
}
