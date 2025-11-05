<?php

namespace App\Services;

use App\Models\Stream;
use App\Models\Server;
use App\Models\User;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\Cache;

class LoadBalancerService
{
    /**
     * Select best server for stream
     */
    public function selectBestServer(Stream $stream, User $user): ?Server
    {
        // Get available servers for this stream
        $servers = $this->getAvailableServers($stream);

        if ($servers->isEmpty()) {
            return null;
        }

        // Check for forced server
        if ($user->forced_server_id) {
            $forcedServer = $servers->firstWhere('id', $user->forced_server_id);
            if ($forcedServer && $forcedServer->hasCapacity()) {
                return $forcedServer;
            }
        }

        // Apply load balancing strategy
        $strategy = config('streaming.load_balancing.strategy', 'least_connections');

        return match ($strategy) {
            'least_connections' => $this->leastConnectionsStrategy($servers),
            'round_robin' => $this->roundRobinStrategy($servers),
            'weighted' => $this->weightedStrategy($servers),
            default => $servers->first(),
        };
    }

    /**
     * Get servers available for stream
     */
    private function getAvailableServers(Stream $stream): Collection
    {
        return $stream->servers()
            ->where('is_online', true)
            ->where('is_enabled', true)
            ->where('timeshift_only', false)
            ->wherePivot('status', 'online')
            ->get()
            ->filter(fn($server) => $server->hasCapacity());
    }

    /**
     * Least connections strategy
     */
    private function leastConnectionsStrategy(Collection $servers): Server
    {
        return $servers->sortBy(function ($server) {
            return $server->getActiveClientsCount();
        })->first();
    }

    /**
     * Round robin strategy
     */
    private function roundRobinStrategy(Collection $servers): Server
    {
        $cacheKey = 'load_balancer:round_robin:index';
        $index = Cache::get($cacheKey, 0);

        $server = $servers->values()->get($index % $servers->count());

        Cache::put($cacheKey, $index + 1, 3600);

        return $server;
    }

    /**
     * Weighted strategy based on server capacity
     */
    private function weightedStrategy(Collection $servers): Server
    {
        return $servers->sortBy(function ($server) {
            // Lower percentage = better choice
            return $server->getLoadPercentage();
        })->first();
    }

    /**
     * Get server load statistics
     */
    public function getServerStats(): array
    {
        $servers = Server::available()->get();

        return $servers->map(function ($server) {
            return [
                'id' => $server->id,
                'name' => $server->name,
                'active_connections' => $server->getActiveClientsCount(),
                'max_connections' => $server->max_clients,
                'load_percentage' => $server->getLoadPercentage(),
                'is_online' => $server->is_online,
            ];
        })->toArray();
    }

    /**
     * Distribute load across servers
     */
    public function rebalanceLoad(): int
    {
        // Get overloaded servers
        $overloadedServers = Server::available()
            ->get()
            ->filter(fn($server) => $server->getLoadPercentage() > 90);

        $movedCount = 0;

        foreach ($overloadedServers as $server) {
            // Get sessions on this server
            $sessions = $server->activeSessions()
                ->orderBy('started_at', 'desc')
                ->limit(10)
                ->get();

            foreach ($sessions as $session) {
                // Find better server
                $betterServer = $this->selectBestServer($session->stream, $session->user);

                if ($betterServer && $betterServer->id !== $server->id) {
                    // Migrate session (implementation depends on architecture)
                    // For now, just close and let user reconnect
                    $session->end();
                    $movedCount++;
                }
            }
        }

        return $movedCount;
    }
}
