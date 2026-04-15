# Smart TV Apps

Native Smart TV applications for the IPTV Platform supporting major TV platforms.

## Supported Platforms

| Platform | Technology | Status |
|----------|-----------|--------|
| **Samsung Tizen** | JavaScript + WebAPIs | Ready |
| **LG webOS** | JavaScript + Enact | Ready |
| **Android TV** | Kotlin + ExoPlayer | Ready |
| **Apple tvOS** | Swift + AVKit | Ready |
| **Roku** | BrightScript | Ready |
| **Amazon Fire TV** | Android (FireOS) | Ready |
| **Vizio SmartCast** | HTML5 + JavaScript | Ready |
| **Hisense VIDAA** | HTML5 | Ready |

## Architecture

```
smart-tv-apps/
├── tizen/              # Samsung Smart TV
├── webos/              # LG Smart TV
├── android-tv/         # Android TV (Kotlin)
├── tvos/               # Apple TV (Swift)
├── roku/               # Roku channel
├── shared/             # Shared assets
│   ├── api/           # Common API client
│   ├── ui/            # UI components
│   └── player/        # Video player
└── README.md
```

## Common Features

All Smart TV apps implement:

- Login/Authentication via QR code or email/password
- Live TV channels with EPG
- VOD movies & series
- Search functionality
- Recommendations
- Watchlist & favorites
- Continue watching
- Multi-language subtitles
- Audio track selection
- Quality selection (Auto/SD/HD/FHD/4K)
- Parental controls
- Settings & profile management

## Key Technology Choices

### Video Player
- **Web platforms** (Tizen, webOS, Vizio): Shaka Player + HLS.js
- **Android TV/Fire TV**: ExoPlayer 2.x with Widevine DRM
- **tvOS**: AVPlayer with FairPlay DRM
- **Roku**: Built-in Video node

### Streaming Protocols
- HLS (Apple)
- DASH (cross-platform)
- LL-HLS (low latency)
- WebRTC (P2P)

### DRM Support
- Widevine (Android, Tizen, webOS)
- FairPlay (Apple TV)
- PlayReady (Microsoft, Tizen)

## Quick Start

### Tizen (Samsung)
```bash
cd tizen
tizen build-web
tizen package -t wgt
tizen install -n IPTV.wgt -t TV-DEVICE
```

### webOS (LG)
```bash
cd webos
ares-package .
ares-install com.iptv.app.ipk
ares-launch com.iptv.app
```

### Android TV
```bash
cd android-tv
./gradlew assembleRelease
adb install app/build/outputs/apk/release/app-release.apk
```

### tvOS
```bash
cd tvos
xcodebuild -project IPTVApp.xcodeproj -scheme IPTVApp build
```

### Roku
```bash
cd roku
make
curl -u rokudev:password http://ROKU_IP/plugin_install --form "mysubmit=Install" --form "archive=@iptv.zip"
```

## API Integration

All apps share the same backend API:

```javascript
// Login
POST /api/auth/login
{ "email": "user@example.com", "password": "password" }

// Get streams
GET /api/streams?category=sports

// Get HLS playlist
GET /api/streams/{id}/playlist.m3u8

// Track viewing
POST /api/analytics/view
{ "stream_id": "123", "duration": 600 }
```

## Performance Optimizations

- **Lazy loading**: Load images/content as needed
- **Image optimization**: WebP format, multiple sizes
- **Virtualized lists**: Render only visible items
- **Prefetching**: Preload adjacent content
- **Cache management**: Aggressive caching with TTL
- **Memory limits**: Respect 200-500MB limits on TVs

## Testing

```bash
# Run tests for each platform
npm test                 # Web platforms
./gradlew test           # Android TV
xcodebuild test          # tvOS
brs test/                # Roku
```

## Submission Guidelines

- **Samsung**: Submit to Samsung Apps Store
- **LG**: Submit to LG Content Store
- **Android TV**: Google Play Store
- **tvOS**: Apple App Store
- **Roku**: Roku Channel Store

Each platform has its own certification process. Allow 1-4 weeks for approval.
