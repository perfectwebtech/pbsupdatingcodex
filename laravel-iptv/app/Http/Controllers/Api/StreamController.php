<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Stream;
use App\Models\Category;
use App\Services\StreamingService;
use Illuminate\Http\Request;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Response;

class StreamController extends Controller
{
    public function __construct(
        private StreamingService $streamingService
    ) {}

    /**
     * Get all streams for authenticated user
     */
    public function index(Request $request): JsonResponse
    {
        $user = $request->user();

        $query = $user->package->streams()->active();

        // Filter by type
        if ($request->has('type')) {
            $query->where('type', $request->input('type'));
        }

        // Filter by category
        if ($request->has('category_id')) {
            $query->where('category_id', $request->input('category_id'));
        }

        // Search
        if ($request->has('search')) {
            $query->where('name', 'LIKE', '%' . $request->input('search') . '%');
        }

        $streams = $query->with('category')
            ->orderBy('order')
            ->paginate(50);

        return response()->json([
            'success' => true,
            'data' => $streams->map(function ($stream) {
                return [
                    'id' => $stream->id,
                    'name' => $stream->name,
                    'type' => $stream->type,
                    'icon_url' => $stream->icon_url,
                    'category' => [
                        'id' => $stream->category->id,
                        'name' => $stream->category->name,
                    ],
                    'has_archive' => $stream->hasArchive(),
                    'metadata' => $stream->metadata,
                ];
            }),
            'pagination' => [
                'total' => $streams->total(),
                'per_page' => $streams->perPage(),
                'current_page' => $streams->currentPage(),
                'last_page' => $streams->lastPage(),
            ],
        ]);
    }

    /**
     * Get single stream
     */
    public function show(Request $request, int $id): JsonResponse
    {
        $user = $request->user();
        $stream = Stream::active()->find($id);

        if (!$stream) {
            return response()->json([
                'success' => false,
                'message' => 'Stream not found',
            ], 404);
        }

        if (!$user->hasAccessToStream($stream)) {
            return response()->json([
                'success' => false,
                'message' => 'Access denied',
            ], 403);
        }

        return response()->json([
            'success' => true,
            'data' => [
                'id' => $stream->id,
                'name' => $stream->name,
                'type' => $stream->type,
                'icon_url' => $stream->icon_url,
                'category' => $stream->category,
                'has_archive' => $stream->hasArchive(),
                'metadata' => $stream->metadata,
                'active_viewers' => $stream->getActiveViewersCount(),
            ],
        ]);
    }

    /**
     * Get categories
     */
    public function categories(Request $request): JsonResponse
    {
        $type = $request->input('type');

        $query = Category::active()
            ->rootCategories()
            ->ordered();

        if ($type) {
            $query->where('type', $type);
        }

        $categories = $query->get();

        return response()->json([
            'success' => true,
            'data' => $categories->map(function ($category) {
                return [
                    'id' => $category->id,
                    'name' => $category->name,
                    'type' => $category->type,
                    'streams_count' => $category->streams()->active()->count(),
                ];
            }),
        ]);
    }

    /**
     * Get stream URL
     */
    public function getUrl(Request $request, int $id): JsonResponse
    {
        $user = $request->user();
        $stream = Stream::active()->find($id);

        if (!$stream) {
            return response()->json([
                'success' => false,
                'message' => 'Stream not found',
            ], 404);
        }

        $container = $request->input('container', 'ts');

        try {
            $result = $this->streamingService->getStream(
                $user,
                $id,
                $container,
                $request->ip(),
                $request->userAgent() ?? ''
            );

            return response()->json([
                'success' => true,
                'data' => $result,
            ]);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'message' => $e->getMessage(),
            ], 400);
        }
    }

    /**
     * Get HLS playlist
     */
    public function playlist(Request $request, int $id): Response
    {
        $user = $request->user();
        $stream = Stream::active()->find($id);

        if (!$stream || !$user->hasAccessToStream($stream)) {
            abort(404);
        }

        try {
            // Get best server
            $server = app(\App\Services\LoadBalancerService::class)
                ->selectBestServer($stream, $user);

            $playlist = $this->streamingService->getHlsPlaylist(
                $stream,
                $server,
                $user
            );

            return response($playlist)
                ->header('Content-Type', 'application/x-mpegurl')
                ->header('Cache-Control', 'no-cache, no-store, must-revalidate');
        } catch (\Exception $e) {
            abort(500, $e->getMessage());
        }
    }

    /**
     * Get HLS segment
     */
    public function segment(Request $request, int $id, string $segment): Response
    {
        $user = $request->user();
        $stream = Stream::active()->find($id);

        if (!$stream || !$user->hasAccessToStream($stream)) {
            abort(404);
        }

        try {
            $result = $this->streamingService->getHlsSegment(
                $segment,
                $user,
                $stream
            );

            return response()
                ->file($result['path'], [
                    'Content-Type' => $result['content_type'],
                    'Content-Length' => $result['size'],
                ]);
        } catch (\Exception $e) {
            abort(404);
        }
    }
}
