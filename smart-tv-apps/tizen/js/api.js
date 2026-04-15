/**
 * IPTV Platform API Client for Smart TV
 */
class APIClient {
    constructor(baseURL) {
        this.baseURL = baseURL;
        this.token = null;
    }

    setToken(token) {
        this.token = token;
    }

    async request(method, endpoint, body = null) {
        const headers = {
            'Content-Type': 'application/json',
            'X-Device-Type': 'smart-tv',
            'X-Platform': 'tizen',
        };

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const options = { method, headers };
        if (body) {
            options.body = JSON.stringify(body);
        }

        const response = await fetch(`${this.baseURL}${endpoint}`, options);

        if (!response.ok) {
            throw new Error(`API Error: ${response.status} ${response.statusText}`);
        }

        return response.json();
    }

    // Auth
    async login(email, password) {
        return this.request('POST', '/api/auth/login', { email, password });
    }

    async logout() {
        return this.request('POST', '/api/auth/logout');
    }

    async refreshToken() {
        return this.request('POST', '/api/auth/refresh');
    }

    // Content
    async getTrending() {
        return this.request('GET', '/api/recommendations/trending');
    }

    async getRecommendations() {
        return this.request('GET', '/api/recommendations/personalized');
    }

    async getNewReleases() {
        return this.request('GET', '/api/content/new-releases');
    }

    async getContinueWatching() {
        return this.request('GET', '/api/users/continue-watching');
    }

    async getCategories() {
        return this.request('GET', '/api/categories');
    }

    async getStreams(categoryId = null) {
        const url = categoryId ? `/api/streams?category=${categoryId}` : '/api/streams';
        return this.request('GET', url);
    }

    async getStreamDetails(id) {
        return this.request('GET', `/api/streams/${id}`);
    }

    async getStreamUrl(id, type) {
        // Get optimal CDN selection
        const cdnSelection = await this.request('GET',
            `/api/cdn/select?content_path=${type}/${id}/playlist.m3u8`);
        return { url: cdnSelection.url, fallbacks: cdnSelection.fallback_urls };
    }

    // Movies & Series
    async getMovies(filters = {}) {
        const params = new URLSearchParams(filters).toString();
        return this.request('GET', `/api/vod/movies?${params}`);
    }

    async getSeries(filters = {}) {
        const params = new URLSearchParams(filters).toString();
        return this.request('GET', `/api/vod/series?${params}`);
    }

    async getEpisodes(seriesId, seasonNumber) {
        return this.request('GET',
            `/api/vod/series/${seriesId}/seasons/${seasonNumber}/episodes`);
    }

    // EPG
    async getEPG(channelId, date) {
        return this.request('GET', `/api/epg/${channelId}?date=${date}`);
    }

    // User actions
    async addToFavorites(contentId, type) {
        return this.request('POST', '/api/users/favorites',
            { content_id: contentId, type });
    }

    async removeFromFavorites(contentId) {
        return this.request('DELETE', `/api/users/favorites/${contentId}`);
    }

    async addToWatchlist(contentId, type) {
        return this.request('POST', '/api/users/watchlist',
            { content_id: contentId, type });
    }

    // Analytics
    async trackView(contentId, type) {
        return this.request('POST', '/api/analytics/view', {
            content_id: contentId,
            type,
            timestamp: Date.now(),
            device: 'smart-tv',
        });
    }

    async updateProgress(contentId, position, duration) {
        return this.request('POST', '/api/analytics/progress', {
            content_id: contentId,
            position,
            duration,
            completed: position / duration > 0.9,
        });
    }

    // Search
    async search(query, type = 'all') {
        return this.request('GET',
            `/api/search?q=${encodeURIComponent(query)}&type=${type}`);
    }

    // Subscription
    async getCurrentSubscription() {
        return this.request('GET', '/api/users/subscription');
    }

    // QR Code Login
    async generateQRCode() {
        return this.request('POST', '/api/auth/qr/generate');
    }

    async checkQRStatus(code) {
        return this.request('GET', `/api/auth/qr/${code}/status`);
    }
}

// Make available globally
window.APIClient = APIClient;
