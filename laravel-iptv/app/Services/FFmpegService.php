<?php

namespace App\Services;

use App\Models\Stream;
use App\Models\Server;
use Illuminate\Support\Facades\Process;
use Illuminate\Support\Facades\Log;

class FFmpegService
{
    private string $ffmpegPath;
    private string $ffprobePath;
    private string $streamsPath;

    public function __construct()
    {
        $this->ffmpegPath = config('streaming.ffmpeg.path');
        $this->ffprobePath = config('streaming.ffmpeg.probe_path');
        $this->streamsPath = config('streaming.paths.streams');
    }

    /**
     * Start stream encoding
     */
    public function startStream(Stream $stream, Server $server): bool
    {
        if (empty($stream->source_urls)) {
            Log::error("No source URLs for stream {$stream->id}");
            return false;
        }

        $sourceUrl = $stream->source_urls[0];
        $outputPath = "{$this->streamsPath}/{$stream->id}_.m3u8";

        $command = $this->buildFFmpegCommand($sourceUrl, $outputPath, $stream);

        try {
            // Start FFmpeg in background
            $process = Process::start($command);

            // Store PID
            $pidFile = "{$this->streamsPath}/{$stream->id}_.pid";
            file_put_contents($pidFile, $process->id());

            // Update stream status
            $stream->servers()->updateExistingPivot($server->id, [
                'pid' => $process->id(),
                'status' => 'online',
            ]);

            Log::info("Started stream {$stream->id} with PID {$process->id()}");

            return true;
        } catch (\Exception $e) {
            Log::error("Failed to start stream {$stream->id}: " . $e->getMessage());
            return false;
        }
    }

    /**
     * Stop stream
     */
    public function stopStream(Stream $stream, Server $server): bool
    {
        $pivot = $stream->servers()
            ->wherePivot('server_id', $server->id)
            ->first()
            ->pivot;

        if (!$pivot || !$pivot->pid) {
            return false;
        }

        try {
            // Kill process
            Process::run("kill -9 {$pivot->pid}");

            // Update status
            $stream->servers()->updateExistingPivot($server->id, [
                'pid' => null,
                'status' => 'offline',
            ]);

            // Clean up files
            $this->cleanupStreamFiles($stream);

            Log::info("Stopped stream {$stream->id}");

            return true;
        } catch (\Exception $e) {
            Log::error("Failed to stop stream {$stream->id}: " . $e->getMessage());
            return false;
        }
    }

    /**
     * Build FFmpeg command
     */
    private function buildFFmpegCommand(
        string $input,
        string $output,
        Stream $stream
    ): string {
        $segmentDuration = config('streaming.stream.segment_duration', 10);
        $playlistSize = config('streaming.stream.playlist_size', 10);

        return sprintf(
            '%s -i "%s" -c:v copy -c:a copy -f hls ' .
            '-hls_time %d -hls_list_size %d -hls_flags delete_segments ' .
            '-hls_segment_filename "%s/%d_%%d.ts" "%s" ' .
            '> /dev/null 2>&1 &',
            $this->ffmpegPath,
            $input,
            $segmentDuration,
            $playlistSize,
            $this->streamsPath,
            $stream->id,
            $output
        );
    }

    /**
     * Get stream information using ffprobe
     */
    public function probeStream(string $url): ?array
    {
        $command = sprintf(
            '%s -v quiet -print_format json -show_format -show_streams "%s"',
            $this->ffprobePath,
            $url
        );

        try {
            $result = Process::run($command);

            if ($result->successful()) {
                return json_decode($result->output(), true);
            }

            return null;
        } catch (\Exception $e) {
            Log::error("FFprobe failed: " . $e->getMessage());
            return null;
        }
    }

    /**
     * Check if stream process is running
     */
    public function isStreamRunning(int $pid): bool
    {
        $result = Process::run("ps -p {$pid}");
        return $result->successful();
    }

    /**
     * Get stream bitrate
     */
    public function getStreamBitrate(string $playlistPath): ?int
    {
        if (!file_exists($playlistPath)) {
            return null;
        }

        $content = file_get_contents($playlistPath);
        $lines = explode("\n", $content);

        $totalSize = 0;
        $totalDuration = 0;

        foreach ($lines as $line) {
            if (preg_match('/^#EXTINF:([\d.]+)/', $line, $matches)) {
                $totalDuration += (float)$matches[1];
            } elseif (preg_match('/\.ts$/', $line)) {
                $segmentPath = dirname($playlistPath) . '/' . trim($line);
                if (file_exists($segmentPath)) {
                    $totalSize += filesize($segmentPath);
                }
            }
        }

        if ($totalDuration > 0) {
            // Convert to kbps
            return (int)(($totalSize * 8) / $totalDuration / 1000);
        }

        return null;
    }

    /**
     * Transcode stream to different quality
     */
    public function transcodeStream(
        string $input,
        string $output,
        string $profile = 'medium'
    ): bool {
        $profiles = config('streaming.transcoding.profiles');

        if (!isset($profiles[$profile])) {
            Log::error("Invalid transcode profile: {$profile}");
            return false;
        }

        $settings = $profiles[$profile];

        $command = sprintf(
            '%s -i "%s" -c:v libx264 -b:v %s -c:a aac -b:a %s ' .
            '-s %s -r %d -f hls -hls_time 10 "%s"',
            $this->ffmpegPath,
            $input,
            $settings['video_bitrate'],
            $settings['audio_bitrate'],
            $settings['resolution'],
            $settings['fps'],
            $output
        );

        try {
            Process::run($command);
            return true;
        } catch (\Exception $e) {
            Log::error("Transcode failed: " . $e->getMessage());
            return false;
        }
    }

    /**
     * Clean up stream files
     */
    private function cleanupStreamFiles(Stream $stream): void
    {
        $pattern = "{$this->streamsPath}/{$stream->id}_*";
        $files = glob($pattern);

        foreach ($files as $file) {
            @unlink($file);
        }
    }
}
