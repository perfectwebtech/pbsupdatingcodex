import React, { useState, useEffect } from 'react';
import {
  SignalIcon,
  MapPinIcon,
  DevicePhoneMobileIcon,
  XMarkIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
  GlobeAltIcon,
  ClockIcon,
  PlayCircleIcon,
} from '@heroicons/react/24/outline';

interface ActiveSession {
  id: number;
  user_id: number;
  username: string;
  device_id: number;
  device_name: string;
  device_type: string;
  stream_id?: number;
  stream_name?: string;
  ip_address: string;
  country?: string;
  city?: string;
  user_agent: string;
  started_at: string;
  last_ping_at: string;
  bandwidth: number; // Mbps
  quality: string; // HD, FHD, 4K
  buffer_health: number; // percentage
  is_buffering: boolean;
  connection_type: string; // wifi, 4g, 5g, ethernet
}

interface SessionStats {
  total_sessions: number;
  concurrent_viewers: number;
  total_bandwidth: number; // Gbps
  avg_session_duration: number; // minutes
  buffering_ratio: number; // percentage
  peak_concurrent: number;
}

interface GeoData {
  country: string;
  sessions: number;
  percentage: number;
}

const Sessions: React.FC = () => {
  const [sessions, setSessions] = useState<ActiveSession[]>([]);
  const [stats, setStats] = useState<SessionStats>({
    total_sessions: 0,
    concurrent_viewers: 0,
    total_bandwidth: 0,
    avg_session_duration: 0,
    buffering_ratio: 0,
    peak_concurrent: 0,
  });
  const [geoData, setGeoData] = useState<GeoData[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterDevice, setFilterDevice] = useState<string>('all');
  const [filterQuality, setFilterQuality] = useState<string>('all');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

  useEffect(() => {
    loadMockData();

    // Auto-refresh every 5 seconds
    if (autoRefresh) {
      const interval = setInterval(() => {
        loadMockData();
        setLastRefresh(new Date());
      }, 5000);

      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  const loadMockData = () => {
    const countries = ['USA', 'UK', 'Germany', 'France', 'Canada', 'Australia', 'Japan', 'Brazil'];
    const cities = ['New York', 'London', 'Berlin', 'Paris', 'Toronto', 'Sydney', 'Tokyo', 'São Paulo'];
    const deviceTypes = ['android', 'ios', 'web', 'mag', 'smart_tv', 'enigma2'];
    const streamNames = ['CNN International', 'ESPN Sports', 'BBC News', 'Discovery', 'HBO Max'];
    const qualities = ['HD', 'FHD', '4K'];
    const connections = ['wifi', '4g', '5g', 'ethernet'];

    const mockSessions: ActiveSession[] = Array.from({ length: 24 }, (_, i) => {
      const started = new Date(Date.now() - Math.random() * 7200000); // 0-2 hours ago
      const lastPing = new Date(Date.now() - Math.random() * 30000); // 0-30 seconds ago

      return {
        id: i + 1,
        user_id: Math.floor(Math.random() * 1000) + 1,
        username: `user_${Math.floor(Math.random() * 1000)}`,
        device_id: Math.floor(Math.random() * 500) + 1,
        device_name: `${deviceTypes[i % deviceTypes.length]}_device_${i}`,
        device_type: deviceTypes[i % deviceTypes.length],
        stream_id: Math.floor(Math.random() * 100) + 1,
        stream_name: streamNames[i % streamNames.length],
        ip_address: `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
        country: countries[i % countries.length],
        city: cities[i % cities.length],
        user_agent: `Mozilla/5.0 (${deviceTypes[i % deviceTypes.length]})`,
        started_at: started.toISOString(),
        last_ping_at: lastPing.toISOString(),
        bandwidth: parseFloat((Math.random() * 15 + 2).toFixed(2)),
        quality: qualities[i % qualities.length],
        buffer_health: Math.floor(Math.random() * 30 + 70), // 70-100%
        is_buffering: Math.random() > 0.9, // 10% chance of buffering
        connection_type: connections[i % connections.length],
      };
    });

    const mockStats: SessionStats = {
      total_sessions: mockSessions.length,
      concurrent_viewers: mockSessions.filter(s => new Date(s.last_ping_at) > new Date(Date.now() - 60000)).length,
      total_bandwidth: parseFloat(mockSessions.reduce((sum, s) => sum + s.bandwidth, 0).toFixed(2)),
      avg_session_duration: Math.floor(mockSessions.reduce((sum, s) => sum + (Date.now() - new Date(s.started_at).getTime()), 0) / mockSessions.length / 60000),
      buffering_ratio: parseFloat(((mockSessions.filter(s => s.is_buffering).length / mockSessions.length) * 100).toFixed(1)),
      peak_concurrent: mockSessions.length + Math.floor(Math.random() * 10),
    };

    // Geo data
    const geoMap = new Map<string, number>();
    mockSessions.forEach(s => {
      if (s.country) {
        geoMap.set(s.country, (geoMap.get(s.country) || 0) + 1);
      }
    });

    const mockGeoData: GeoData[] = Array.from(geoMap.entries())
      .map(([country, count]) => ({
        country,
        sessions: count,
        percentage: parseFloat(((count / mockSessions.length) * 100).toFixed(1)),
      }))
      .sort((a, b) => b.sessions - a.sessions);

    setSessions(mockSessions);
    setStats(mockStats);
    setGeoData(mockGeoData);
  };

  const handleTerminateSession = (sessionId: number) => {
    if (!confirm('Terminate this session?')) return;

    console.log('Terminating session:', sessionId);
    setSessions(sessions.filter(s => s.id !== sessionId));
    alert('Session terminated successfully!');
  };

  const handleTerminateAll = () => {
    if (!confirm(`Terminate all ${sessions.length} active sessions? This action cannot be undone.`)) return;

    console.log('Terminating all sessions');
    setSessions([]);
    alert('All sessions terminated!');
  };

  const formatDuration = (startedAt: string): string => {
    const start = new Date(startedAt);
    const now = new Date();
    const diff = now.getTime() - start.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    }
    return `${minutes}m`;
  };

  const formatLastSeen = (lastPingAt: string): string => {
    const lastPing = new Date(lastPingAt);
    const now = new Date();
    const diff = Math.floor((now.getTime() - lastPing.getTime()) / 1000);

    if (diff < 10) return 'Just now';
    if (diff < 60) return `${diff}s ago`;
    const minutes = Math.floor(diff / 60);
    return `${minutes}m ago`;
  };

  const getDeviceIcon = (type: string) => {
    const icons: Record<string, string> = {
      android: '🤖',
      ios: '📱',
      web: '💻',
      mag: '📺',
      smart_tv: '📺',
      enigma2: '📡',
    };
    return icons[type] || '📱';
  };

  const getConnectionIcon = (type: string) => {
    const icons: Record<string, string> = {
      wifi: '📶',
      '4g': '📱',
      '5g': '🚀',
      ethernet: '🔌',
    };
    return icons[type] || '📶';
  };

  const getQualityColor = (quality: string): string => {
    const colors: Record<string, string> = {
      'HD': 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
      'FHD': 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
      '4K': 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
    };
    return colors[quality] || colors['HD'];
  };

  const filteredSessions = sessions.filter(session => {
    const matchesSearch = session.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
      session.ip_address.includes(searchTerm) ||
      session.stream_name?.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesDevice = filterDevice === 'all' || session.device_type === filterDevice;
    const matchesQuality = filterQuality === 'all' || session.quality === filterQuality;

    return matchesSearch && matchesDevice && matchesQuality;
  });

  return (
    <div className="p-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
      {/* Header */}
      <div className="mb-6 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <SignalIcon className="w-8 h-8 text-blue-500" />
            Live Sessions Monitor
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Real-time streaming session monitoring and management
          </p>
        </div>

        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="w-4 h-4 text-blue-600 border-gray-300 rounded"
            />
            Auto-refresh
          </label>
          <button
            onClick={() => {
              loadMockData();
              setLastRefresh(new Date());
            }}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg flex items-center gap-2 transition-colors"
          >
            <ArrowPathIcon className="w-5 h-5" />
            Refresh
          </button>
          <button
            onClick={handleTerminateAll}
            className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg flex items-center gap-2 transition-colors"
          >
            <XMarkIcon className="w-5 h-5" />
            Terminate All
          </button>
        </div>
      </div>

      {/* Last Refresh */}
      <div className="mb-4 text-sm text-gray-600 dark:text-gray-400">
        Last updated: {lastRefresh.toLocaleTimeString()}
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4 mb-8">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-500 dark:text-gray-400">Active Sessions</span>
            <SignalIcon className="w-6 h-6 text-green-500" />
          </div>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total_sessions}</p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-500 dark:text-gray-400">Concurrent</span>
            <PlayCircleIcon className="w-6 h-6 text-blue-500" />
          </div>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.concurrent_viewers}</p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-500 dark:text-gray-400">Bandwidth</span>
            <div className="text-2xl">🌐</div>
          </div>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total_bandwidth} Gbps</p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-500 dark:text-gray-400">Avg Duration</span>
            <ClockIcon className="w-6 h-6 text-purple-500" />
          </div>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.avg_session_duration}m</p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-500 dark:text-gray-400">Buffering</span>
            <div className="text-2xl">⏸️</div>
          </div>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.buffering_ratio}%</p>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-500 dark:text-gray-400">Peak Today</span>
            <div className="text-2xl">🔥</div>
          </div>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{stats.peak_concurrent}</p>
        </div>
      </div>

      {/* Geographic Distribution */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-8">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
          <GlobeAltIcon className="w-6 h-6 text-blue-500" />
          Geographic Distribution
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
          {geoData.map((geo, index) => (
            <div key={index} className="text-center">
              <div className="text-2xl mb-1">{geo.country === 'USA' ? '🇺🇸' : geo.country === 'UK' ? '🇬🇧' : '🌍'}</div>
              <div className="text-sm font-medium text-gray-900 dark:text-white">{geo.country}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">{geo.sessions} ({geo.percentage}%)</div>
            </div>
          ))}
        </div>
      </div>

      {/* Toolbar */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <div className="flex flex-col lg:flex-row gap-4 items-center justify-between">
          <div className="flex flex-col sm:flex-row gap-4 w-full lg:w-auto">
            {/* Search */}
            <div className="relative flex-1 lg:w-64">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search user, IP, stream..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              />
            </div>

            {/* Device Filter */}
            <select
              value={filterDevice}
              onChange={(e) => setFilterDevice(e.target.value)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="all">All Devices</option>
              <option value="android">Android</option>
              <option value="ios">iOS</option>
              <option value="web">Web</option>
              <option value="mag">MAG</option>
              <option value="smart_tv">Smart TV</option>
              <option value="enigma2">Enigma2</option>
            </select>

            {/* Quality Filter */}
            <select
              value={filterQuality}
              onChange={(e) => setFilterQuality(e.target.value)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="all">All Quality</option>
              <option value="HD">HD</option>
              <option value="FHD">Full HD</option>
              <option value="4K">4K</option>
            </select>
          </div>

          <div className="text-sm text-gray-600 dark:text-gray-400">
            Showing {filteredSessions.length} of {sessions.length} sessions
          </div>
        </div>
      </div>

      {/* Sessions Table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">User / Device</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Stream</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Location</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Quality</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Duration</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {filteredSessions.map((session) => (
                <tr key={session.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <span className="text-2xl">{getDeviceIcon(session.device_type)}</span>
                      <div>
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {session.username}
                        </div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">
                          {session.device_name}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-900 dark:text-white">
                      {session.stream_name || 'Not watching'}
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      {session.bandwidth} Mbps
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <MapPinIcon className="w-4 h-4 text-gray-400" />
                      <div>
                        <div className="text-sm text-gray-900 dark:text-white">{session.country}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">{session.ip_address}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="space-y-1">
                      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getQualityColor(session.quality)}`}>
                        {session.quality}
                      </span>
                      <div className="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1">
                        {getConnectionIcon(session.connection_type)} {session.connection_type}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-900 dark:text-white">
                      {formatDuration(session.started_at)}
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      {formatLastSeen(session.last_ping_at)}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    {session.is_buffering ? (
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
                        ⏸️ Buffering
                      </span>
                    ) : (
                      <div className="space-y-1">
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                          ▶️ Streaming
                        </span>
                        <div className="text-xs text-gray-500 dark:text-gray-400">
                          Buffer: {session.buffer_health}%
                        </div>
                      </div>
                    )}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => handleTerminateSession(session.id)}
                      className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                      title="Terminate Session"
                    >
                      <XMarkIcon className="w-5 h-5" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Empty State */}
      {filteredSessions.length === 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-12 text-center mt-4">
          <SignalIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-500 dark:text-gray-400 text-lg">No active sessions</p>
        </div>
      )}
    </div>
  );
};

export default Sessions;
