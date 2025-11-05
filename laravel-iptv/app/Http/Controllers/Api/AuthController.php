<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Services\AuthenticationService;
use Illuminate\Http\Request;
use Illuminate\Http\JsonResponse;
use Illuminate\Support\Facades\Validator;

class AuthController extends Controller
{
    public function __construct(
        private AuthenticationService $authService
    ) {}

    /**
     * User login
     */
    public function login(Request $request): JsonResponse
    {
        $validator = Validator::make($request->all(), [
            'username' => 'required|string|max:50',
            'password' => 'required|string',
        ]);

        if ($validator->fails()) {
            return response()->json([
                'success' => false,
                'errors' => $validator->errors(),
            ], 422);
        }

        try {
            $result = $this->authService->authenticate(
                $request->input('username'),
                $request->input('password'),
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
            ], 401);
        }
    }

    /**
     * Refresh token
     */
    public function refresh(Request $request): JsonResponse
    {
        try {
            $token = $request->bearerToken();

            if (!$token) {
                return response()->json([
                    'success' => false,
                    'message' => 'Token not provided',
                ], 401);
            }

            $result = $this->authService->refreshToken($token);

            return response()->json([
                'success' => true,
                'data' => $result,
            ]);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'message' => $e->getMessage(),
            ], 401);
        }
    }

    /**
     * Logout
     */
    public function logout(Request $request): JsonResponse
    {
        try {
            $token = $request->bearerToken();

            if ($token) {
                $this->authService->logout($token);
            }

            return response()->json([
                'success' => true,
                'message' => 'Logged out successfully',
            ]);
        } catch (\Exception $e) {
            return response()->json([
                'success' => true, // Still return success even if token invalid
                'message' => 'Logged out',
            ]);
        }
    }

    /**
     * Get authenticated user info
     */
    public function me(Request $request): JsonResponse
    {
        $user = $request->user();

        return response()->json([
            'success' => true,
            'data' => [
                'id' => $user->id,
                'username' => $user->username,
                'email' => $user->email,
                'max_connections' => $user->max_connections,
                'active_connections' => $user->getActiveConnectionsCount(),
                'is_trial' => $user->is_trial,
                'expires_at' => $user->expires_at?->toIso8601String(),
                'package' => $user->package ? [
                    'id' => $user->package->id,
                    'name' => $user->package->name,
                ] : null,
            ],
        ]);
    }
}
