/**
 * IPTV Platform - Smart TV App (Samsung Tizen)
 * Main application entry point
 */

(function() {
    'use strict';

    const API_BASE = 'https://api.iptv.example.com';
    let currentFocus = null;
    let userToken = null;

    // Initialize app
    function init() {
        console.log('IPTV TV App initializing...');

        // Setup remote control
        setupRemoteControl();

        // Initialize API client
        const api = new APIClient(API_BASE);

        // Check authentication
        userToken = localStorage.getItem('iptv_token');
        if (!userToken) {
            showLoginScreen();
        } else {
            api.setToken(userToken);
            loadHomeScreen(api);
        }
    }

    // Setup TV remote control handling
    function setupRemoteControl() {
        // Register key codes for Samsung TV
        if (window.tizen) {
            try {
                tizen.tvinputdevice.registerKey('MediaPlayPause');
                tizen.tvinputdevice.registerKey('MediaPlay');
                tizen.tvinputdevice.registerKey('MediaPause');
                tizen.tvinputdevice.registerKey('MediaStop');
                tizen.tvinputdevice.registerKey('MediaRewind');
                tizen.tvinputdevice.registerKey('MediaFastForward');
                tizen.tvinputdevice.registerKey('ChannelUp');
                tizen.tvinputdevice.registerKey('ChannelDown');
                tizen.tvinputdevice.registerKey('VolumeUp');
                tizen.tvinputdevice.registerKey('VolumeDown');
                tizen.tvinputdevice.registerKey('Mute');
                tizen.tvinputdevice.registerKey('Info');
                tizen.tvinputdevice.registerKey('Menu');
                tizen.tvinputdevice.registerKey('ColorF0Red');
                tizen.tvinputdevice.registerKey('ColorF1Green');
                tizen.tvinputdevice.registerKey('ColorF2Yellow');
                tizen.tvinputdevice.registerKey('ColorF3Blue');
            } catch (e) {
                console.error('Failed to register TV keys:', e);
            }
        }

        // Key event handler
        document.addEventListener('keydown', handleKeyPress);
    }

    function handleKeyPress(e) {
        switch(e.keyCode) {
            case 37: // LEFT
                navigate('left');
                break;
            case 38: // UP
                navigate('up');
                break;
            case 39: // RIGHT
                navigate('right');
                break;
            case 40: // DOWN
                navigate('down');
                break;
            case 13: // ENTER/OK
                activate();
                break;
            case 10009: // BACK (Tizen)
            case 27: // ESC
                goBack();
                break;
            case 415: // PLAY
                playPause();
                break;
            case 19: // PAUSE
                playPause();
                break;
            case 412: // REWIND
                seek(-10);
                break;
            case 417: // FAST FORWARD
                seek(10);
                break;
        }
    }

    // D-pad navigation
    function navigate(direction) {
        const focusables = document.querySelectorAll('[tabindex]:not([tabindex="-1"]), button, a');
        const currentIndex = Array.from(focusables).indexOf(document.activeElement);

        if (currentIndex === -1) {
            focusables[0]?.focus();
            return;
        }

        let nextIndex = currentIndex;

        // Calculate next focused element based on direction
        // (simplified - in production, use spatial navigation)
        switch(direction) {
            case 'left':
                nextIndex = Math.max(0, currentIndex - 1);
                break;
            case 'right':
                nextIndex = Math.min(focusables.length - 1, currentIndex + 1);
                break;
            case 'up':
                nextIndex = Math.max(0, currentIndex - 5);
                break;
            case 'down':
                nextIndex = Math.min(focusables.length - 1, currentIndex + 5);
                break;
        }

        focusables[nextIndex]?.focus();
    }

    function activate() {
        if (document.activeElement) {
            document.activeElement.click();
        }
    }

    function goBack() {
        const overlay = document.getElementById('player-overlay');
        if (!overlay.classList.contains('hidden')) {
            overlay.classList.add('hidden');
            const player = document.getElementById('video-player');
            player.pause();
            player.src = '';
        } else {
            // Confirm exit
            if (confirm('Exit IPTV?')) {
                if (window.tizen) {
                    tizen.application.getCurrentApplication().exit();
                }
            }
        }
    }

    function playPause() {
        const player = document.getElementById('video-player');
        if (player.paused) {
            player.play();
        } else {
            player.pause();
        }
    }

    function seek(seconds) {
        const player = document.getElementById('video-player');
        player.currentTime = Math.max(0, player.currentTime + seconds);
    }

    // Login screen
    function showLoginScreen() {
        document.body.innerHTML = `
            <div class="login-screen">
                <div class="login-box">
                    <h1>Welcome to IPTV Platform</h1>
                    <p>Sign in to your account</p>
                    <input type="email" id="email" placeholder="Email" tabindex="1">
                    <input type="password" id="password" placeholder="Password" tabindex="2">
                    <button onclick="login()" tabindex="3">Sign In</button>
                    <p class="qr-hint">Or scan QR code with mobile app</p>
                    <div id="qr-code"></div>
                </div>
            </div>
        `;
        document.getElementById('email').focus();
    }

    window.login = async function() {
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;

        try {
            const api = new APIClient(API_BASE);
            const response = await api.login(email, password);

            userToken = response.token;
            localStorage.setItem('iptv_token', userToken);
            api.setToken(userToken);

            loadHomeScreen(api);
        } catch (e) {
            alert('Login failed: ' + e.message);
        }
    };

    // Load home screen
    async function loadHomeScreen(api) {
        try {
            const [trending, recommended, newReleases, continueWatching] = await Promise.all([
                api.getTrending(),
                api.getRecommendations(),
                api.getNewReleases(),
                api.getContinueWatching(),
            ]);

            renderRow('trending', trending.items);
            renderRow('recommended', recommended.items);
            renderRow('new-releases', newReleases.items);
            renderRow('continue-watching', continueWatching.items);

            // Set hero
            if (trending.items?.length > 0) {
                setHero(trending.items[0]);
            }
        } catch (e) {
            console.error('Failed to load home:', e);
        }
    }

    function renderRow(containerId, items) {
        const container = document.getElementById(containerId);
        if (!container || !items) return;

        container.innerHTML = items.map((item, idx) => `
            <div class="content-card" tabindex="${idx + 100}" data-id="${item.id}" data-type="${item.type}">
                <img src="${item.poster_url}" alt="${item.title}" loading="lazy">
                <div class="card-info">
                    <h3>${item.title}</h3>
                    <span class="rating">⭐ ${item.rating || 'N/A'}</span>
                </div>
            </div>
        `).join('');

        // Add click handlers
        container.querySelectorAll('.content-card').forEach(card => {
            card.addEventListener('click', () => playContent(card.dataset.id, card.dataset.type));
        });
    }

    function setHero(item) {
        document.getElementById('hero-title').textContent = item.title;
        document.getElementById('hero-description').textContent = item.description;
        document.getElementById('hero-play').onclick = () => playContent(item.id, item.type);
    }

    async function playContent(id, type) {
        const api = new APIClient(API_BASE);
        api.setToken(userToken);

        try {
            const stream = await api.getStreamUrl(id, type);

            const overlay = document.getElementById('player-overlay');
            const player = document.getElementById('video-player');

            overlay.classList.remove('hidden');

            // Use HLS.js for HLS playback if available
            if (window.Hls && Hls.isSupported() && stream.url.includes('.m3u8')) {
                const hls = new Hls();
                hls.loadSource(stream.url);
                hls.attachMedia(player);
                hls.on(Hls.Events.MANIFEST_PARSED, () => player.play());
            } else {
                player.src = stream.url;
                player.play();
            }

            // Track viewing
            api.trackView(id, type);
        } catch (e) {
            alert('Failed to play: ' + e.message);
        }
    }

    // Initialize when ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
