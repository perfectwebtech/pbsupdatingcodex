import React, { useState, useEffect } from 'react';
import {
  DevicePhoneMobileIcon,
  TvIcon,
  ComputerDesktopIcon,
  DeviceTabletIcon,
  SignalIcon,
  LockClosedIcon,
  LockOpenIcon,
  TrashIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
} from '@heroicons/react/24/outline';

// Device types
type DeviceType = 'mag' | 'enigma2' | 'android' | 'ios' | 'web' | 'stb' | 'smart_tv';

interface Device {
  id: number;
  user_id: number;
  device_name: string;
  device_type: DeviceType;
  mac_address: string;
  ip_address?: string;
  user_agent?: string;
  model?: string;
  os_version?: string;
  app_version?: string;
  is_active: boolean;
  is_blocked: boolean;
  last_seen_at?: string;
  activated_at?: string;
  blocked_at?: string;
  blocked_reason?: string;
  created_at: string;
  updated_at: string;
}

interface DeviceSession {
  id: number;
  device_id: number;
  user_id: number;
  stream_id?: number;
  ip_address: string;
  user_agent: string;
  started_at: string;
  last_ping_at: string;
  ended_at?: string;
  duration?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface DeviceStats {
  total_devices: number;
  active_devices: number;
  blocked_devices: number;
  devices_by_type: { [key: string]: number };
  recent_devices: number;
  active_sessions: number;
  total_sessions: number;
}

interface DeviceFormData {
  user_id: number;
  device_name: string;
  device_type: DeviceType;
  mac_address: string;
  ip_address?: string;
  user_agent?: string;
  model?: string;
  os_version?: string;
  app_version?: string;
}

const Devices: React.FC = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [sessions, setSessions] = useState<DeviceSession[]>([]);
  const [stats, setStats] = useState<DeviceStats>({
    total_devices: 0,
    active_devices: 0,
    blocked_devices: 0,
    devices_by_type: {},
    recent_devices: 0,
    active_sessions: 0,
    total_sessions: 0,
  });

  const [viewMode, setViewMode] = useState<'devices' | 'sessions'>('devices');
  const [showRegisterModal, setShowRegisterModal] = useState(false);
  const [showBlockModal, setShowBlockModal] = useState(false);
  const [showSessionsModal, setShowSessionsModal] = useState(false);
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
  const [selectedDeviceSessions, setSelectedDeviceSessions] = useState<DeviceSession[]>([]);
  const [blockReason, setBlockReason] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState<DeviceType | 'all'>('all');
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'blocked'>('all');

  const [formData, setFormData] = useState<DeviceFormData>({
    user_id: 1,
    device_name: '',
    device_type: 'android',
    mac_address: '',
    ip_address: '',
    user_agent: '',
    model: '',
    os_version: '',
    app_version: '',
  });

  // Mock data for demonstration
  useEffect(() => {
    loadMockData();
  }, []);

  const loadMockData = () => {
    // Mock devices
    const mockDevices: Device[] = [
      {
        id: 1,
        user_id: 1,
        device_name: 'John\'s Android TV',
        device_type: 'android',
        mac_address: 'AA:BB:CC:DD:EE:01',
        ip_address: '192.168.1.100',
        user_agent: 'AndroidTV/10.0',
        model: 'Sony Bravia X90J',
        os_version: 'Android 10',
        app_version: '1.5.2',
        is_active: true,
        is_blocked: false,
        last_seen_at: new Date(Date.now() - 30 * 60000).toISOString(),
        activated_at: new Date(Date.now() - 30 * 86400000).toISOString(),
        created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 30 * 60000).toISOString(),
      },
      {
        id: 2,
        user_id: 1,
        device_name: 'Living Room MAG',
        device_type: 'mag',
        mac_address: 'AA:BB:CC:DD:EE:02',
        ip_address: '192.168.1.101',
        user_agent: 'MAG254/v1.0',
        model: 'MAG254',
        os_version: 'Linux 3.3',
        is_active: true,
        is_blocked: false,
        last_seen_at: new Date(Date.now() - 120 * 60000).toISOString(),
        activated_at: new Date(Date.now() - 60 * 86400000).toISOString(),
        created_at: new Date(Date.now() - 60 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 120 * 60000).toISOString(),
      },
      {
        id: 3,
        user_id: 2,
        device_name: 'Sarah\'s iPhone',
        device_type: 'ios',
        mac_address: 'AA:BB:CC:DD:EE:03',
        ip_address: '192.168.1.102',
        user_agent: 'iOS/16.0',
        model: 'iPhone 14 Pro',
        os_version: 'iOS 16.0',
        app_version: '2.1.0',
        is_active: true,
        is_blocked: false,
        last_seen_at: new Date(Date.now() - 10 * 60000).toISOString(),
        activated_at: new Date(Date.now() - 15 * 86400000).toISOString(),
        created_at: new Date(Date.now() - 15 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 10 * 60000).toISOString(),
      },
      {
        id: 4,
        user_id: 3,
        device_name: 'Bedroom Smart TV',
        device_type: 'smart_tv',
        mac_address: 'AA:BB:CC:DD:EE:04',
        ip_address: '192.168.1.103',
        user_agent: 'Tizen/6.0',
        model: 'Samsung QN90A',
        os_version: 'Tizen 6.0',
        is_active: true,
        is_blocked: true,
        last_seen_at: new Date(Date.now() - 2 * 86400000).toISOString(),
        activated_at: new Date(Date.now() - 45 * 86400000).toISOString(),
        blocked_at: new Date(Date.now() - 1 * 86400000).toISOString(),
        blocked_reason: 'Suspected sharing violation',
        created_at: new Date(Date.now() - 45 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 1 * 86400000).toISOString(),
      },
      {
        id: 5,
        user_id: 4,
        device_name: 'Office Enigma2',
        device_type: 'enigma2',
        mac_address: 'AA:BB:CC:DD:EE:05',
        ip_address: '192.168.1.104',
        user_agent: 'Enigma2/v1.0',
        model: 'VU+ Uno 4K',
        os_version: 'OpenPLi 7.0',
        is_active: true,
        is_blocked: false,
        last_seen_at: new Date(Date.now() - 5 * 60000).toISOString(),
        activated_at: new Date(Date.now() - 90 * 86400000).toISOString(),
        created_at: new Date(Date.now() - 90 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 5 * 60000).toISOString(),
      },
      {
        id: 6,
        user_id: 5,
        device_name: 'Web Browser (Chrome)',
        device_type: 'web',
        mac_address: 'AA:BB:CC:DD:EE:06',
        ip_address: '192.168.1.105',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0',
        model: 'Desktop Browser',
        os_version: 'Windows 11',
        is_active: true,
        is_blocked: false,
        last_seen_at: new Date(Date.now() - 1 * 60000).toISOString(),
        activated_at: new Date(Date.now() - 7 * 86400000).toISOString(),
        created_at: new Date(Date.now() - 7 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 1 * 60000).toISOString(),
      },
    ];

    // Mock sessions
    const mockSessions: DeviceSession[] = [
      {
        id: 1,
        device_id: 1,
        user_id: 1,
        stream_id: 101,
        ip_address: '192.168.1.100',
        user_agent: 'AndroidTV/10.0',
        started_at: new Date(Date.now() - 45 * 60000).toISOString(),
        last_ping_at: new Date(Date.now() - 1 * 60000).toISOString(),
        is_active: true,
        created_at: new Date(Date.now() - 45 * 60000).toISOString(),
        updated_at: new Date(Date.now() - 1 * 60000).toISOString(),
      },
      {
        id: 2,
        device_id: 3,
        user_id: 2,
        stream_id: 102,
        ip_address: '192.168.1.102',
        user_agent: 'iOS/16.0',
        started_at: new Date(Date.now() - 20 * 60000).toISOString(),
        last_ping_at: new Date(Date.now() - 30000).toISOString(),
        is_active: true,
        created_at: new Date(Date.now() - 20 * 60000).toISOString(),
        updated_at: new Date(Date.now() - 30000).toISOString(),
      },
      {
        id: 3,
        device_id: 5,
        user_id: 4,
        stream_id: 103,
        ip_address: '192.168.1.104',
        user_agent: 'Enigma2/v1.0',
        started_at: new Date(Date.now() - 15 * 60000).toISOString(),
        last_ping_at: new Date(Date.now() - 15000).toISOString(),
        is_active: true,
        created_at: new Date(Date.now() - 15 * 60000).toISOString(),
        updated_at: new Date(Date.now() - 15000).toISOString(),
      },
      {
        id: 4,
        device_id: 6,
        user_id: 5,
        stream_id: 104,
        ip_address: '192.168.1.105',
        user_agent: 'Chrome/120.0',
        started_at: new Date(Date.now() - 5 * 60000).toISOString(),
        last_ping_at: new Date(Date.now() - 5000).toISOString(),
        is_active: true,
        created_at: new Date(Date.now() - 5 * 60000).toISOString(),
        updated_at: new Date(Date.now() - 5000).toISOString(),
      },
    ];

    // Mock stats
    const mockStats: DeviceStats = {
      total_devices: mockDevices.length,
      active_devices: mockDevices.filter(d => !d.is_blocked && d.last_seen_at && new Date(d.last_seen_at) > new Date(Date.now() - 24 * 3600000)).length,
      blocked_devices: mockDevices.filter(d => d.is_blocked).length,
      devices_by_type: {
        android: mockDevices.filter(d => d.device_type === 'android').length,
        mag: mockDevices.filter(d => d.device_type === 'mag').length,
        ios: mockDevices.filter(d => d.device_type === 'ios').length,
        smart_tv: mockDevices.filter(d => d.device_type === 'smart_tv').length,
        enigma2: mockDevices.filter(d => d.device_type === 'enigma2').length,
        web: mockDevices.filter(d => d.device_type === 'web').length,
        stb: mockDevices.filter(d => d.device_type === 'stb').length,
      },
      recent_devices: mockDevices.filter(d => new Date(d.created_at) > new Date(Date.now() - 7 * 86400000)).length,
      active_sessions: mockSessions.filter(s => s.is_active).length,
      total_sessions: mockSessions.length,
    };

    setDevices(mockDevices);
    setSessions(mockSessions);
    setStats(mockStats);
  };

  const handleRegisterDevice = async () => {
    if (!formData.device_name || !formData.mac_address) {
      alert('Please fill in all required fields');
      return;
    }

    // In production: await deviceAPI.registerDevice(formData);
    console.log('Registering device:', formData);

    // Mock: Add to list
    const newDevice: Device = {
      id: devices.length + 1,
      ...formData,
      is_active: true,
      is_blocked: false,
      activated_at: new Date().toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setDevices([newDevice, ...devices]);

    setShowRegisterModal(false);
    resetForm();
    alert('Device registered successfully!');
  };

  const handleBlockDevice = async () => {
    if (!selectedDevice || !blockReason) {
      alert('Please provide a reason for blocking');
      return;
    }

    // In production: await deviceAPI.blockDevice(selectedDevice.id, { reason: blockReason });
    console.log('Blocking device:', selectedDevice.id, blockReason);

    // Mock: Update device
    setDevices(devices.map(d =>
      d.id === selectedDevice.id
        ? { ...d, is_blocked: true, blocked_at: new Date().toISOString(), blocked_reason: blockReason }
        : d
    ));

    setShowBlockModal(false);
    setBlockReason('');
    setSelectedDevice(null);
    alert('Device blocked successfully!');
  };

  const handleUnblockDevice = async (device: Device) => {
    if (!confirm(`Unblock device "${device.device_name}"?`)) return;

    // In production: await deviceAPI.unblockDevice(device.id);
    console.log('Unblocking device:', device.id);

    // Mock: Update device
    setDevices(devices.map(d =>
      d.id === device.id
        ? { ...d, is_blocked: false, blocked_at: undefined, blocked_reason: undefined }
        : d
    ));

    alert('Device unblocked successfully!');
  };

  const handleDeleteDevice = async (device: Device) => {
    if (!confirm(`Delete device "${device.device_name}"? This will terminate all active sessions.`)) return;

    // In production: await deviceAPI.deleteDevice(device.id);
    console.log('Deleting device:', device.id);

    // Mock: Remove device
    setDevices(devices.filter(d => d.id !== device.id));
    alert('Device deleted successfully!');
  };

  const handleViewSessions = async (device: Device) => {
    // In production: const sessions = await deviceAPI.getDeviceSessions(device.id);
    console.log('Loading sessions for device:', device.id);

    // Mock: Filter sessions for this device
    const deviceSessions = sessions.filter(s => s.device_id === device.id);
    setSelectedDeviceSessions(deviceSessions);
    setSelectedDevice(device);
    setShowSessionsModal(true);
  };

  const handleTerminateSession = async (session: DeviceSession) => {
    if (!confirm('Terminate this session?')) return;

    // In production: await deviceAPI.terminateSession(session.id);
    console.log('Terminating session:', session.id);

    // Mock: Update session
    setSessions(sessions.map(s =>
      s.id === session.id
        ? { ...s, is_active: false, ended_at: new Date().toISOString() }
        : s
    ));
    setSelectedDeviceSessions(selectedDeviceSessions.map(s =>
      s.id === session.id
        ? { ...s, is_active: false, ended_at: new Date().toISOString() }
        : s
    ));

    alert('Session terminated successfully!');
  };

  const resetForm = () => {
    setFormData({
      user_id: 1,
      device_name: '',
      device_type: 'android',
      mac_address: '',
      ip_address: '',
      user_agent: '',
      model: '',
      os_version: '',
      app_version: '',
    });
  };

  const getDeviceIcon = (type: DeviceType) => {
    switch (type) {
      case 'android':
      case 'ios':
        return <DevicePhoneMobileIcon className="w-5 h-5" />;
      case 'mag':
      case 'enigma2':
      case 'stb':
        return <TvIcon className="w-5 h-5" />;
      case 'smart_tv':
        return <TvIcon className="w-5 h-5" />;
      case 'web':
        return <ComputerDesktopIcon className="w-5 h-5" />;
      default:
        return <DeviceTabletIcon className="w-5 h-5" />;
    }
  };

  const getDeviceTypeName = (type: DeviceType): string => {
    const names: Record<DeviceType, string> = {
      mag: 'MAG Box',
      enigma2: 'Enigma2',
      android: 'Android',
      ios: 'iOS',
      web: 'Web Browser',
      stb: 'Set-Top Box',
      smart_tv: 'Smart TV',
    };
    return names[type];
  };

  const formatTimeAgo = (dateString?: string): string => {
    if (!dateString) return 'Never';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const filteredDevices = devices.filter(device => {
    const matchesSearch = device.device_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      device.mac_address.toLowerCase().includes(searchTerm.toLowerCase()) ||
      device.model?.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesType = filterType === 'all' || device.device_type === filterType;

    const matchesStatus = filterStatus === 'all' ||
      (filterStatus === 'blocked' && device.is_blocked) ||
      (filterStatus === 'active' && !device.is_blocked);

    return matchesSearch && matchesType && matchesStatus;
  });

  return (
    <div className="p-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <DevicePhoneMobileIcon className="w-8 h-8 text-blue-500" />
          Device Management
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          Manage devices, track sessions, and monitor device activity
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Total Devices</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-2">{stats.total_devices}</p>
            </div>
            <DevicePhoneMobileIcon className="w-12 h-12 text-blue-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Active Devices</p>
              <p className="text-3xl font-bold text-green-600 dark:text-green-400 mt-2">{stats.active_devices}</p>
            </div>
            <SignalIcon className="w-12 h-12 text-green-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Blocked Devices</p>
              <p className="text-3xl font-bold text-red-600 dark:text-red-400 mt-2">{stats.blocked_devices}</p>
            </div>
            <LockClosedIcon className="w-12 h-12 text-red-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Active Sessions</p>
              <p className="text-3xl font-bold text-purple-600 dark:text-purple-400 mt-2">{stats.active_sessions}</p>
            </div>
            <TvIcon className="w-12 h-12 text-purple-500" />
          </div>
        </div>
      </div>

      {/* View Mode Tabs */}
      <div className="mb-6 border-b border-gray-200 dark:border-gray-700">
        <div className="flex gap-4">
          <button
            onClick={() => setViewMode('devices')}
            className={`pb-4 px-2 border-b-2 font-medium transition-colors ${
              viewMode === 'devices'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            Devices ({devices.length})
          </button>
          <button
            onClick={() => setViewMode('sessions')}
            className={`pb-4 px-2 border-b-2 font-medium transition-colors ${
              viewMode === 'sessions'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            Active Sessions ({sessions.filter(s => s.is_active).length})
          </button>
        </div>
      </div>

      {/* Devices View */}
      {viewMode === 'devices' && (
        <>
          {/* Toolbar */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
            <div className="flex flex-col lg:flex-row gap-4 items-center justify-between">
              <div className="flex flex-col sm:flex-row gap-4 w-full lg:w-auto">
                {/* Search */}
                <div className="relative flex-1 lg:w-64">
                  <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                  <input
                    type="text"
                    placeholder="Search devices..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                {/* Type Filter */}
                <select
                  value={filterType}
                  onChange={(e) => setFilterType(e.target.value as DeviceType | 'all')}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="all">All Types</option>
                  <option value="android">Android</option>
                  <option value="ios">iOS</option>
                  <option value="mag">MAG Box</option>
                  <option value="enigma2">Enigma2</option>
                  <option value="smart_tv">Smart TV</option>
                  <option value="web">Web</option>
                  <option value="stb">STB</option>
                </select>

                {/* Status Filter */}
                <select
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value as 'all' | 'active' | 'blocked')}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="all">All Status</option>
                  <option value="active">Active</option>
                  <option value="blocked">Blocked</option>
                </select>
              </div>

              <button
                onClick={() => setShowRegisterModal(true)}
                className="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg flex items-center gap-2 justify-center transition-colors"
              >
                <PlusIcon className="w-5 h-5" />
                Register Device
              </button>
            </div>
          </div>

          {/* Devices Table */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Device
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      MAC Address
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Last Seen
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  {filteredDevices.map((device) => (
                    <tr key={device.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-3">
                          <div className="text-gray-500 dark:text-gray-400">
                            {getDeviceIcon(device.device_type)}
                          </div>
                          <div>
                            <div className="text-sm font-medium text-gray-900 dark:text-white">
                              {device.device_name}
                            </div>
                            <div className="text-xs text-gray-500 dark:text-gray-400">
                              {device.model || 'No model info'}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-900 dark:text-white">
                          {getDeviceTypeName(device.device_type)}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm font-mono text-gray-900 dark:text-white">
                          {device.mac_address}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-500 dark:text-gray-400">
                          {formatTimeAgo(device.last_seen_at)}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {device.is_blocked ? (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
                            <LockClosedIcon className="w-3 h-3" />
                            Blocked
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                            <SignalIcon className="w-3 h-3" />
                            Active
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => handleViewSessions(device)}
                            className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                            title="View Sessions"
                          >
                            <TvIcon className="w-5 h-5" />
                          </button>
                          {device.is_blocked ? (
                            <button
                              onClick={() => handleUnblockDevice(device)}
                              className="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300"
                              title="Unblock"
                            >
                              <LockOpenIcon className="w-5 h-5" />
                            </button>
                          ) : (
                            <button
                              onClick={() => {
                                setSelectedDevice(device);
                                setShowBlockModal(true);
                              }}
                              className="text-orange-600 hover:text-orange-900 dark:text-orange-400 dark:hover:text-orange-300"
                              title="Block"
                            >
                              <LockClosedIcon className="w-5 h-5" />
                            </button>
                          )}
                          <button
                            onClick={() => handleDeleteDevice(device)}
                            className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                            title="Delete"
                          >
                            <TrashIcon className="w-5 h-5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {/* Sessions View */}
      {viewMode === 'sessions' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Device
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Stream ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    IP Address
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Started
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Last Ping
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {sessions.filter(s => s.is_active).map((session) => {
                  const device = devices.find(d => d.id === session.device_id);
                  return (
                    <tr key={session.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {device?.device_name || `Device #${session.device_id}`}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-900 dark:text-white">
                          {session.stream_id || 'N/A'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm font-mono text-gray-900 dark:text-white">
                          {session.ip_address}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-500 dark:text-gray-400">
                          {formatTimeAgo(session.started_at)}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-500 dark:text-gray-400">
                          {formatTimeAgo(session.last_ping_at)}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button
                          onClick={() => handleTerminateSession(session)}
                          className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                        >
                          Terminate
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Register Device Modal */}
      {showRegisterModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Register New Device</h2>

              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Device Name *
                    </label>
                    <input
                      type="text"
                      value={formData.device_name}
                      onChange={(e) => setFormData({ ...formData, device_name: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="e.g., Living Room TV"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Device Type *
                    </label>
                    <select
                      value={formData.device_type}
                      onChange={(e) => setFormData({ ...formData, device_type: e.target.value as DeviceType })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    >
                      <option value="android">Android</option>
                      <option value="ios">iOS</option>
                      <option value="mag">MAG Box</option>
                      <option value="enigma2">Enigma2</option>
                      <option value="smart_tv">Smart TV</option>
                      <option value="web">Web Browser</option>
                      <option value="stb">Set-Top Box</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    MAC Address * (XX:XX:XX:XX:XX:XX)
                  </label>
                  <input
                    type="text"
                    value={formData.mac_address}
                    onChange={(e) => setFormData({ ...formData, mac_address: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono"
                    placeholder="AA:BB:CC:DD:EE:FF"
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Model
                    </label>
                    <input
                      type="text"
                      value={formData.model}
                      onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="e.g., Samsung QN90A"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      OS Version
                    </label>
                    <input
                      type="text"
                      value={formData.os_version}
                      onChange={(e) => setFormData({ ...formData, os_version: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="e.g., Android 12"
                    />
                  </div>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={handleRegisterDevice}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Register Device
                </button>
                <button
                  onClick={() => {
                    setShowRegisterModal(false);
                    resetForm();
                  }}
                  className="flex-1 bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Block Device Modal */}
      {showBlockModal && selectedDevice && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full">
            <div className="p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">Block Device</h2>
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                Block device: <strong>{selectedDevice.device_name}</strong>
              </p>

              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Reason *
                </label>
                <textarea
                  value={blockReason}
                  onChange={(e) => setBlockReason(e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  placeholder="Enter reason for blocking this device..."
                />
              </div>

              <div className="flex gap-3">
                <button
                  onClick={handleBlockDevice}
                  className="flex-1 bg-red-600 hover:bg-red-700 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Block Device
                </button>
                <button
                  onClick={() => {
                    setShowBlockModal(false);
                    setBlockReason('');
                    setSelectedDevice(null);
                  }}
                  className="flex-1 bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Device Sessions Modal */}
      {showSessionsModal && selectedDevice && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
                Sessions: {selectedDevice.device_name}
              </h2>

              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50 dark:bg-gray-700">
                    <tr>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                        Stream ID
                      </th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                        IP Address
                      </th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                        Started
                      </th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                        Status
                      </th>
                      <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                    {selectedDeviceSessions.length === 0 ? (
                      <tr>
                        <td colSpan={5} className="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                          No sessions found for this device
                        </td>
                      </tr>
                    ) : (
                      selectedDeviceSessions.map((session) => (
                        <tr key={session.id}>
                          <td className="px-4 py-3 text-sm text-gray-900 dark:text-white">
                            {session.stream_id || 'N/A'}
                          </td>
                          <td className="px-4 py-3 text-sm font-mono text-gray-900 dark:text-white">
                            {session.ip_address}
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                            {formatTimeAgo(session.started_at)}
                          </td>
                          <td className="px-4 py-3">
                            {session.is_active ? (
                              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                                Active
                              </span>
                            ) : (
                              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300">
                                Ended
                              </span>
                            )}
                          </td>
                          <td className="px-4 py-3 text-right">
                            {session.is_active && (
                              <button
                                onClick={() => handleTerminateSession(session)}
                                className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 text-sm"
                              >
                                Terminate
                              </button>
                            )}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="mt-6">
                <button
                  onClick={() => {
                    setShowSessionsModal(false);
                    setSelectedDevice(null);
                    setSelectedDeviceSessions([]);
                  }}
                  className="w-full bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Devices;
