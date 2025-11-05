<?php

namespace App\Services;

use GeoIp2\Database\Reader;
use GeoIp2\Exception\AddressNotFoundException;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Log;

class GeoIPService
{
    private Reader $reader;

    public function __construct()
    {
        $dbPath = config('streaming.geoip.database_path');

        if (!file_exists($dbPath)) {
            Log::warning("GeoIP database not found at: {$dbPath}");
        } else {
            $this->reader = new Reader($dbPath);
        }
    }

    /**
     * Get country code from IP
     */
    public function getCountryCode(string $ipAddress): ?string
    {
        if (!isset($this->reader)) {
            return null;
        }

        $cacheKey = "geoip:country:{$ipAddress}";
        $cacheDuration = config('streaming.geoip.cache_duration', 86400);

        return Cache::remember($cacheKey, $cacheDuration, function () use ($ipAddress) {
            try {
                $record = $this->reader->country($ipAddress);
                return $record->country->isoCode;
            } catch (AddressNotFoundException $e) {
                Log::debug("IP not found in GeoIP database: {$ipAddress}");
                return null;
            } catch (\Exception $e) {
                Log::error("GeoIP lookup failed: " . $e->getMessage());
                return null;
            }
        });
    }

    /**
     * Get full location data
     */
    public function getLocation(string $ipAddress): ?array
    {
        if (!isset($this->reader)) {
            return null;
        }

        $cacheKey = "geoip:location:{$ipAddress}";
        $cacheDuration = config('streaming.geoip.cache_duration', 86400);

        return Cache::remember($cacheKey, $cacheDuration, function () use ($ipAddress) {
            try {
                $record = $this->reader->city($ipAddress);

                return [
                    'country_code' => $record->country->isoCode,
                    'country_name' => $record->country->name,
                    'city' => $record->city->name,
                    'postal_code' => $record->postal->code,
                    'latitude' => $record->location->latitude,
                    'longitude' => $record->location->longitude,
                    'timezone' => $record->location->timeZone,
                ];
            } catch (AddressNotFoundException $e) {
                return null;
            } catch (\Exception $e) {
                Log::error("GeoIP location lookup failed: " . $e->getMessage());
                return null;
            }
        });
    }

    /**
     * Get ISP information
     */
    public function getISP(string $ipAddress): ?string
    {
        // This would require GeoIP2 ISP database
        // For now, return null or implement external API call
        return null;
    }

    /**
     * Check if IP is from server/datacenter
     */
    public function isServerIP(string $ipAddress): bool
    {
        // This would require ASN database or external service
        // Placeholder implementation
        return false;
    }

    /**
     * Bulk lookup for multiple IPs
     */
    public function bulkLookup(array $ipAddresses): array
    {
        $results = [];

        foreach ($ipAddresses as $ip) {
            $results[$ip] = $this->getLocation($ip);
        }

        return $results;
    }
}
