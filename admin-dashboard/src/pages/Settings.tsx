import React, { useState, useEffect } from 'react';
import {
  Cog6ToothIcon,
  ServerIcon,
  EnvelopeIcon,
  CreditCardIcon,
  CloudIcon,
  FilmIcon,
  ShieldCheckIcon,
  WrenchScrewdriverIcon,
  KeyIcon,
  BellAlertIcon,
} from '@heroicons/react/24/outline';

// Interfaces
interface GeneralSettings {
  siteName: string;
  siteUrl: string;
  timezone: string;
  dateFormat: string;
  timeFormat: string;
  currency: string;
  logo: string;
  favicon: string;
  defaultLanguage: string;
  allowRegistration: boolean;
  requireEmailVerification: boolean;
}

interface SMTPSettings {
  host: string;
  port: number;
  username: string;
  password: string;
  fromAddress: string;
  fromName: string;
  encryption: 'none' | 'ssl' | 'tls';
  enabled: boolean;
}

interface PaymentSettings {
  stripe: {
    enabled: boolean;
    publicKey: string;
    secretKey: string;
    webhookSecret: string;
  };
  paypal: {
    enabled: boolean;
    clientId: string;
    clientSecret: string;
    mode: 'sandbox' | 'live';
  };
  razorpay: {
    enabled: boolean;
    keyId: string;
    keySecret: string;
  };
}

interface CDNSettings {
  enabled: boolean;
  provider: 'cloudflare' | 'amazon' | 'bunny' | 'custom';
  url: string;
  apiKey: string;
  zoneId: string;
  pullZone: string;
  storageZone: string;
}

interface FFmpegSettings {
  path: string;
  workers: number;
  timeout: number;
  hardware_acceleration: boolean;
  encoder: 'libx264' | 'h264_nvenc' | 'h264_qsv' | 'h264_videotoolbox';
  presets: {
    sd: string;
    hd: string;
    fhd: string;
    uhd: string;
  };
  audio_codec: string;
  segment_duration: number;
}

interface BackupSettings {
  enabled: boolean;
  schedule: 'daily' | 'weekly' | 'monthly';
  time: string;
  retention: number;
  location: 'local' | 's3' | 'ftp';
  s3Bucket?: string;
  s3Region?: string;
  s3AccessKey?: string;
  s3SecretKey?: string;
  ftpHost?: string;
  ftpUsername?: string;
  ftpPassword?: string;
}

interface SecuritySettings {
  rateLimiting: {
    enabled: boolean;
    maxRequests: number;
    windowMinutes: number;
  };
  ipWhitelist: {
    enabled: boolean;
    ips: string[];
  };
  sessionTimeout: number;
  passwordPolicy: {
    minLength: number;
    requireUppercase: boolean;
    requireLowercase: boolean;
    requireNumbers: boolean;
    requireSpecialChars: boolean;
  };
  twoFactorAuth: boolean;
  captcha: {
    enabled: boolean;
    siteKey: string;
    secretKey: string;
  };
}

interface MaintenanceSettings {
  enabled: boolean;
  message: string;
  allowedIPs: string[];
  startTime?: string;
  endTime?: string;
}

interface LicenseSettings {
  key: string;
  status: 'active' | 'expired' | 'invalid';
  expiresAt: string;
  domain: string;
  maxUsers: number;
  maxStreams: number;
}

interface AdvancedSettings {
  debugMode: boolean;
  logLevel: 'debug' | 'info' | 'warning' | 'error';
  cacheEnabled: boolean;
  cacheTTL: number;
  queueDriver: 'redis' | 'database' | 'sync';
  sessionDriver: 'file' | 'cookie' | 'database' | 'redis';
  broadcastDriver: 'pusher' | 'redis' | 'log';
}

type TabType = 'general' | 'smtp' | 'payment' | 'cdn' | 'ffmpeg' | 'backup' | 'security' | 'maintenance' | 'license' | 'advanced';

const Settings: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabType>('general');
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  // State for all settings
  const [generalSettings, setGeneralSettings] = useState<GeneralSettings>({
    siteName: 'IPTV Platform Pro',
    siteUrl: 'https://iptv.example.com',
    timezone: 'UTC',
    dateFormat: 'Y-m-d',
    timeFormat: '24h',
    currency: 'USD',
    logo: '/logo.png',
    favicon: '/favicon.ico',
    defaultLanguage: 'en',
    allowRegistration: true,
    requireEmailVerification: true,
  });

  const [smtpSettings, setSmtpSettings] = useState<SMTPSettings>({
    host: 'smtp.gmail.com',
    port: 587,
    username: '',
    password: '',
    fromAddress: 'noreply@example.com',
    fromName: 'IPTV Platform',
    encryption: 'tls',
    enabled: false,
  });

  const [paymentSettings, setPaymentSettings] = useState<PaymentSettings>({
    stripe: {
      enabled: false,
      publicKey: '',
      secretKey: '',
      webhookSecret: '',
    },
    paypal: {
      enabled: false,
      clientId: '',
      clientSecret: '',
      mode: 'sandbox',
    },
    razorpay: {
      enabled: false,
      keyId: '',
      keySecret: '',
    },
  });

  const [cdnSettings, setCdnSettings] = useState<CDNSettings>({
    enabled: false,
    provider: 'cloudflare',
    url: '',
    apiKey: '',
    zoneId: '',
    pullZone: '',
    storageZone: '',
  });

  const [ffmpegSettings, setFfmpegSettings] = useState<FFmpegSettings>({
    path: '/usr/bin/ffmpeg',
    workers: 4,
    timeout: 3600,
    hardware_acceleration: false,
    encoder: 'libx264',
    presets: {
      sd: '-c:v libx264 -preset fast -b:v 800k -s 640x360',
      hd: '-c:v libx264 -preset medium -b:v 2500k -s 1280x720',
      fhd: '-c:v libx264 -preset medium -b:v 5000k -s 1920x1080',
      uhd: '-c:v libx264 -preset slow -b:v 15000k -s 3840x2160',
    },
    audio_codec: 'aac',
    segment_duration: 6,
  });

  const [backupSettings, setBackupSettings] = useState<BackupSettings>({
    enabled: true,
    schedule: 'daily',
    time: '02:00',
    retention: 7,
    location: 'local',
  });

  const [securitySettings, setSecuritySettings] = useState<SecuritySettings>({
    rateLimiting: {
      enabled: true,
      maxRequests: 60,
      windowMinutes: 1,
    },
    ipWhitelist: {
      enabled: false,
      ips: [],
    },
    sessionTimeout: 120,
    passwordPolicy: {
      minLength: 8,
      requireUppercase: true,
      requireLowercase: true,
      requireNumbers: true,
      requireSpecialChars: false,
    },
    twoFactorAuth: false,
    captcha: {
      enabled: false,
      siteKey: '',
      secretKey: '',
    },
  });

  const [maintenanceSettings, setMaintenanceSettings] = useState<MaintenanceSettings>({
    enabled: false,
    message: 'We are currently performing system maintenance. Please check back soon.',
    allowedIPs: [],
  });

  const [licenseSettings, setLicenseSettings] = useState<LicenseSettings>({
    key: 'IPTV-PRO-XXXX-XXXX-XXXX-XXXX',
    status: 'active',
    expiresAt: '2025-12-31',
    domain: 'iptv.example.com',
    maxUsers: 10000,
    maxStreams: 1000,
  });

  const [advancedSettings, setAdvancedSettings] = useState<AdvancedSettings>({
    debugMode: false,
    logLevel: 'info',
    cacheEnabled: true,
    cacheTTL: 3600,
    queueDriver: 'redis',
    sessionDriver: 'redis',
    broadcastDriver: 'redis',
  });

  const tabs = [
    { id: 'general' as TabType, name: 'General', icon: Cog6ToothIcon },
    { id: 'smtp' as TabType, name: 'Email/SMTP', icon: EnvelopeIcon },
    { id: 'payment' as TabType, name: 'Payments', icon: CreditCardIcon },
    { id: 'cdn' as TabType, name: 'CDN', icon: CloudIcon },
    { id: 'ffmpeg' as TabType, name: 'FFmpeg', icon: FilmIcon },
    { id: 'backup' as TabType, name: 'Backup', icon: ServerIcon },
    { id: 'security' as TabType, name: 'Security', icon: ShieldCheckIcon },
    { id: 'maintenance' as TabType, name: 'Maintenance', icon: WrenchScrewdriverIcon },
    { id: 'license' as TabType, name: 'License', icon: KeyIcon },
    { id: 'advanced' as TabType, name: 'Advanced', icon: BellAlertIcon },
  ];

  const handleSave = async () => {
    setSaving(true);

    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1500));

    // Here we would send the settings to the backend
    console.log('Saving settings:', {
      general: generalSettings,
      smtp: smtpSettings,
      payment: paymentSettings,
      cdn: cdnSettings,
      ffmpeg: ffmpegSettings,
      backup: backupSettings,
      security: securitySettings,
      maintenance: maintenanceSettings,
      license: licenseSettings,
      advanced: advancedSettings,
    });

    setSaving(false);
    setSaved(true);

    setTimeout(() => setSaved(false), 3000);
  };

  const handleTestSMTP = async () => {
    alert('Sending test email...');
    // Simulate test email
    await new Promise(resolve => setTimeout(resolve, 1000));
    alert('Test email sent successfully!');
  };

  const handleTestFFmpeg = async () => {
    alert('Testing FFmpeg installation...');
    // Simulate FFmpeg test
    await new Promise(resolve => setTimeout(resolve, 1000));
    alert('FFmpeg is working correctly!\nVersion: 5.1.2\nCodecs: H.264, H.265, VP9, AAC');
  };

  const handleBackupNow = async () => {
    alert('Starting backup process...');
    // Simulate backup
    await new Promise(resolve => setTimeout(resolve, 2000));
    alert('Backup completed successfully!\nSize: 245 MB\nLocation: /backups/backup-2025-11-06.sql.gz');
  };

  const handleClearCache = async () => {
    alert('Clearing cache...');
    // Simulate cache clear
    await new Promise(resolve => setTimeout(resolve, 1000));
    alert('Cache cleared successfully!\nCleared: 1.2 GB');
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          System Settings
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Configure your IPTV platform settings and preferences
        </p>
      </div>

      {/* Tabs */}
      <div className="mb-6 border-b border-gray-200 dark:border-gray-700">
        <div className="flex flex-wrap gap-2 -mb-px">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-4 py-2 border-b-2 font-medium text-sm transition-colors ${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
                }`}
              >
                <Icon className="w-5 h-5" />
                {tab.name}
              </button>
            );
          })}
        </div>
      </div>

      {/* Tab Content */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
        {/* General Settings */}
        {activeTab === 'general' && (
          <div className="space-y-6">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              General Settings
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Site Name
                </label>
                <input
                  type="text"
                  value={generalSettings.siteName}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, siteName: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Site URL
                </label>
                <input
                  type="url"
                  value={generalSettings.siteUrl}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, siteUrl: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Timezone
                </label>
                <select
                  value={generalSettings.timezone}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, timezone: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="UTC">UTC</option>
                  <option value="America/New_York">America/New York</option>
                  <option value="America/Los_Angeles">America/Los Angeles</option>
                  <option value="Europe/London">Europe/London</option>
                  <option value="Europe/Paris">Europe/Paris</option>
                  <option value="Asia/Dubai">Asia/Dubai</option>
                  <option value="Asia/Tokyo">Asia/Tokyo</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Currency
                </label>
                <select
                  value={generalSettings.currency}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, currency: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="USD">USD - US Dollar</option>
                  <option value="EUR">EUR - Euro</option>
                  <option value="GBP">GBP - British Pound</option>
                  <option value="AED">AED - UAE Dirham</option>
                  <option value="INR">INR - Indian Rupee</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Date Format
                </label>
                <select
                  value={generalSettings.dateFormat}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, dateFormat: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="Y-m-d">YYYY-MM-DD</option>
                  <option value="m/d/Y">MM/DD/YYYY</option>
                  <option value="d/m/Y">DD/MM/YYYY</option>
                  <option value="d-M-Y">DD-Mon-YYYY</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Time Format
                </label>
                <select
                  value={generalSettings.timeFormat}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, timeFormat: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="24h">24 Hour</option>
                  <option value="12h">12 Hour (AM/PM)</option>
                </select>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={generalSettings.allowRegistration}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, allowRegistration: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">Allow User Registration</span>
              </label>

              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={generalSettings.requireEmailVerification}
                  onChange={(e) => setGeneralSettings({ ...generalSettings, requireEmailVerification: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">Require Email Verification</span>
              </label>
            </div>
          </div>
        )}

        {/* SMTP Settings */}
        {activeTab === 'smtp' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                Email / SMTP Settings
              </h2>
              <button
                onClick={handleTestSMTP}
                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
              >
                Send Test Email
              </button>
            </div>

            <div className="mb-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={smtpSettings.enabled}
                  onChange={(e) => setSmtpSettings({ ...smtpSettings, enabled: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Enable SMTP</span>
              </label>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  SMTP Host
                </label>
                <input
                  type="text"
                  value={smtpSettings.host}
                  onChange={(e) => setSmtpSettings({ ...smtpSettings, host: e.target.value })}
                  placeholder="smtp.gmail.com"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  SMTP Port
                </label>
                <input
                  type="number"
                  value={smtpSettings.port}
                  onChange={(e) => setSmtpSettings({ ...smtpSettings, port: parseInt(e.target.value) })}
                  placeholder="587"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Username
                </label>
                <input
                  type="text"
                  value={smtpSettings.username}
                  onChange={(e) => setSmtpSettings({ ...smtpSettings, username: e.target.value })}
                  placeholder="your-email@gmail.com"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Password
                </label>
                <input
                  type="password"
                  value={smtpSettings.password}
                  onChange={(e) => setSmtpSettings({ ...smtpSettings, password: e.target.value })}
                  placeholder="••••••••"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  From Address
                </label>
                <input
                  type="email"
                  value={smtpSettings.fromAddress}
                  onChange={(e) => setSmtpSettings({ ...smtpSettings, fromAddress: e.target.value })}
                  placeholder="noreply@example.com"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  From Name
                </label>
                <input
                  type="text"
                  value={smtpSettings.fromName}
                  onChange={(e) => setSmtpSettings({ ...smtpSettings, fromName: e.target.value })}
                  placeholder="IPTV Platform"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Encryption
                </label>
                <div className="flex gap-4">
                  {(['none', 'ssl', 'tls'] as const).map((enc) => (
                    <label key={enc} className="flex items-center gap-2">
                      <input
                        type="radio"
                        checked={smtpSettings.encryption === enc}
                        onChange={() => setSmtpSettings({ ...smtpSettings, encryption: enc })}
                        className="w-4 h-4 text-blue-600"
                      />
                      <span className="text-sm text-gray-700 dark:text-gray-300">{enc.toUpperCase()}</span>
                    </label>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Payment Settings */}
        {activeTab === 'payment' && (
          <div className="space-y-8">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Payment Gateway Settings
            </h2>

            {/* Stripe */}
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Stripe</h3>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={paymentSettings.stripe.enabled}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      stripe: { ...paymentSettings.stripe, enabled: e.target.checked }
                    })}
                    className="w-4 h-4 text-blue-600 rounded"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-300">Enabled</span>
                </label>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Publishable Key
                  </label>
                  <input
                    type="text"
                    value={paymentSettings.stripe.publicKey}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      stripe: { ...paymentSettings.stripe, publicKey: e.target.value }
                    })}
                    placeholder="pk_test_..."
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Secret Key
                  </label>
                  <input
                    type="password"
                    value={paymentSettings.stripe.secretKey}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      stripe: { ...paymentSettings.stripe, secretKey: e.target.value }
                    })}
                    placeholder="sk_test_..."
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Webhook Secret
                  </label>
                  <input
                    type="password"
                    value={paymentSettings.stripe.webhookSecret}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      stripe: { ...paymentSettings.stripe, webhookSecret: e.target.value }
                    })}
                    placeholder="whsec_..."
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>
              </div>
            </div>

            {/* PayPal */}
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">PayPal</h3>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={paymentSettings.paypal.enabled}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      paypal: { ...paymentSettings.paypal, enabled: e.target.checked }
                    })}
                    className="w-4 h-4 text-blue-600 rounded"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-300">Enabled</span>
                </label>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Client ID
                  </label>
                  <input
                    type="text"
                    value={paymentSettings.paypal.clientId}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      paypal: { ...paymentSettings.paypal, clientId: e.target.value }
                    })}
                    placeholder="AYxxx..."
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Client Secret
                  </label>
                  <input
                    type="password"
                    value={paymentSettings.paypal.clientSecret}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      paypal: { ...paymentSettings.paypal, clientSecret: e.target.value }
                    })}
                    placeholder="ELxxx..."
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Mode
                  </label>
                  <div className="flex gap-4">
                    {(['sandbox', 'live'] as const).map((mode) => (
                      <label key={mode} className="flex items-center gap-2">
                        <input
                          type="radio"
                          checked={paymentSettings.paypal.mode === mode}
                          onChange={() => setPaymentSettings({
                            ...paymentSettings,
                            paypal: { ...paymentSettings.paypal, mode }
                          })}
                          className="w-4 h-4 text-blue-600"
                        />
                        <span className="text-sm text-gray-700 dark:text-gray-300 capitalize">{mode}</span>
                      </label>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Razorpay */}
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Razorpay</h3>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={paymentSettings.razorpay.enabled}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      razorpay: { ...paymentSettings.razorpay, enabled: e.target.checked }
                    })}
                    className="w-4 h-4 text-blue-600 rounded"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-300">Enabled</span>
                </label>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Key ID
                  </label>
                  <input
                    type="text"
                    value={paymentSettings.razorpay.keyId}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      razorpay: { ...paymentSettings.razorpay, keyId: e.target.value }
                    })}
                    placeholder="rzp_test_..."
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Key Secret
                  </label>
                  <input
                    type="password"
                    value={paymentSettings.razorpay.keySecret}
                    onChange={(e) => setPaymentSettings({
                      ...paymentSettings,
                      razorpay: { ...paymentSettings.razorpay, keySecret: e.target.value }
                    })}
                    placeholder="••••••••"
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* CDN Settings */}
        {activeTab === 'cdn' && (
          <div className="space-y-6">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              CDN Settings
            </h2>

            <div className="mb-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={cdnSettings.enabled}
                  onChange={(e) => setCdnSettings({ ...cdnSettings, enabled: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Enable CDN</span>
              </label>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  CDN Provider
                </label>
                <select
                  value={cdnSettings.provider}
                  onChange={(e) => setCdnSettings({ ...cdnSettings, provider: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="cloudflare">Cloudflare</option>
                  <option value="amazon">Amazon CloudFront</option>
                  <option value="bunny">Bunny CDN</option>
                  <option value="custom">Custom CDN</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  CDN URL
                </label>
                <input
                  type="url"
                  value={cdnSettings.url}
                  onChange={(e) => setCdnSettings({ ...cdnSettings, url: e.target.value })}
                  placeholder="https://cdn.example.com"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  API Key
                </label>
                <input
                  type="password"
                  value={cdnSettings.apiKey}
                  onChange={(e) => setCdnSettings({ ...cdnSettings, apiKey: e.target.value })}
                  placeholder="••••••••"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Zone ID
                </label>
                <input
                  type="text"
                  value={cdnSettings.zoneId}
                  onChange={(e) => setCdnSettings({ ...cdnSettings, zoneId: e.target.value })}
                  placeholder="abc123..."
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Pull Zone
                </label>
                <input
                  type="text"
                  value={cdnSettings.pullZone}
                  onChange={(e) => setCdnSettings({ ...cdnSettings, pullZone: e.target.value })}
                  placeholder="mycdn-pull"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Storage Zone
                </label>
                <input
                  type="text"
                  value={cdnSettings.storageZone}
                  onChange={(e) => setCdnSettings({ ...cdnSettings, storageZone: e.target.value })}
                  placeholder="mycdn-storage"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>
            </div>
          </div>
        )}

        {/* FFmpeg Settings */}
        {activeTab === 'ffmpeg' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                FFmpeg Transcoding Settings
              </h2>
              <button
                onClick={handleTestFFmpeg}
                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
              >
                Test FFmpeg
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  FFmpeg Path
                </label>
                <input
                  type="text"
                  value={ffmpegSettings.path}
                  onChange={(e) => setFfmpegSettings({ ...ffmpegSettings, path: e.target.value })}
                  placeholder="/usr/bin/ffmpeg"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Worker Threads
                </label>
                <input
                  type="number"
                  value={ffmpegSettings.workers}
                  onChange={(e) => setFfmpegSettings({ ...ffmpegSettings, workers: parseInt(e.target.value) })}
                  min="1"
                  max="16"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Timeout (seconds)
                </label>
                <input
                  type="number"
                  value={ffmpegSettings.timeout}
                  onChange={(e) => setFfmpegSettings({ ...ffmpegSettings, timeout: parseInt(e.target.value) })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Video Encoder
                </label>
                <select
                  value={ffmpegSettings.encoder}
                  onChange={(e) => setFfmpegSettings({ ...ffmpegSettings, encoder: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="libx264">libx264 (CPU)</option>
                  <option value="h264_nvenc">NVIDIA NVENC</option>
                  <option value="h264_qsv">Intel Quick Sync</option>
                  <option value="h264_videotoolbox">Apple VideoToolbox</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Audio Codec
                </label>
                <input
                  type="text"
                  value={ffmpegSettings.audio_codec}
                  onChange={(e) => setFfmpegSettings({ ...ffmpegSettings, audio_codec: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Segment Duration (seconds)
                </label>
                <input
                  type="number"
                  value={ffmpegSettings.segment_duration}
                  onChange={(e) => setFfmpegSettings({ ...ffmpegSettings, segment_duration: parseInt(e.target.value) })}
                  min="2"
                  max="10"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>
            </div>

            <div className="mb-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={ffmpegSettings.hardware_acceleration}
                  onChange={(e) => setFfmpegSettings({ ...ffmpegSettings, hardware_acceleration: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Enable Hardware Acceleration</span>
              </label>
            </div>

            {/* Quality Presets */}
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Quality Presets</h3>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  SD (360p) Preset
                </label>
                <input
                  type="text"
                  value={ffmpegSettings.presets.sd}
                  onChange={(e) => setFfmpegSettings({
                    ...ffmpegSettings,
                    presets: { ...ffmpegSettings.presets, sd: e.target.value }
                  })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono text-xs"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  HD (720p) Preset
                </label>
                <input
                  type="text"
                  value={ffmpegSettings.presets.hd}
                  onChange={(e) => setFfmpegSettings({
                    ...ffmpegSettings,
                    presets: { ...ffmpegSettings.presets, hd: e.target.value }
                  })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono text-xs"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Full HD (1080p) Preset
                </label>
                <input
                  type="text"
                  value={ffmpegSettings.presets.fhd}
                  onChange={(e) => setFfmpegSettings({
                    ...ffmpegSettings,
                    presets: { ...ffmpegSettings.presets, fhd: e.target.value }
                  })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono text-xs"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  4K (2160p) Preset
                </label>
                <input
                  type="text"
                  value={ffmpegSettings.presets.uhd}
                  onChange={(e) => setFfmpegSettings({
                    ...ffmpegSettings,
                    presets: { ...ffmpegSettings.presets, uhd: e.target.value }
                  })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono text-xs"
                />
              </div>
            </div>
          </div>
        )}

        {/* Backup Settings */}
        {activeTab === 'backup' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                Backup Settings
              </h2>
              <button
                onClick={handleBackupNow}
                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
              >
                Backup Now
              </button>
            </div>

            <div className="mb-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={backupSettings.enabled}
                  onChange={(e) => setBackupSettings({ ...backupSettings, enabled: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Enable Automatic Backups</span>
              </label>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Backup Schedule
                </label>
                <select
                  value={backupSettings.schedule}
                  onChange={(e) => setBackupSettings({ ...backupSettings, schedule: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Backup Time
                </label>
                <input
                  type="time"
                  value={backupSettings.time}
                  onChange={(e) => setBackupSettings({ ...backupSettings, time: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Retention Days
                </label>
                <input
                  type="number"
                  value={backupSettings.retention}
                  onChange={(e) => setBackupSettings({ ...backupSettings, retention: parseInt(e.target.value) })}
                  min="1"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Backup Location
                </label>
                <select
                  value={backupSettings.location}
                  onChange={(e) => setBackupSettings({ ...backupSettings, location: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="local">Local Storage</option>
                  <option value="s3">Amazon S3</option>
                  <option value="ftp">FTP Server</option>
                </select>
              </div>
            </div>

            {/* S3 Settings */}
            {backupSettings.location === 's3' && (
              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-4">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Amazon S3 Configuration</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      S3 Bucket
                    </label>
                    <input
                      type="text"
                      placeholder="my-backup-bucket"
                      className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      S3 Region
                    </label>
                    <input
                      type="text"
                      placeholder="us-east-1"
                      className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    />
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Security Settings */}
        {activeTab === 'security' && (
          <div className="space-y-8">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Security Settings
            </h2>

            {/* Rate Limiting */}
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Rate Limiting</h3>
                <label className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={securitySettings.rateLimiting.enabled}
                    onChange={(e) => setSecuritySettings({
                      ...securitySettings,
                      rateLimiting: { ...securitySettings.rateLimiting, enabled: e.target.checked }
                    })}
                    className="w-4 h-4 text-blue-600 rounded"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-300">Enabled</span>
                </label>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Max Requests
                  </label>
                  <input
                    type="number"
                    value={securitySettings.rateLimiting.maxRequests}
                    onChange={(e) => setSecuritySettings({
                      ...securitySettings,
                      rateLimiting: { ...securitySettings.rateLimiting, maxRequests: parseInt(e.target.value) }
                    })}
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Time Window (minutes)
                  </label>
                  <input
                    type="number"
                    value={securitySettings.rateLimiting.windowMinutes}
                    onChange={(e) => setSecuritySettings({
                      ...securitySettings,
                      rateLimiting: { ...securitySettings.rateLimiting, windowMinutes: parseInt(e.target.value) }
                    })}
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>
              </div>
            </div>

            {/* Password Policy */}
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Password Policy</h3>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Minimum Length
                  </label>
                  <input
                    type="number"
                    value={securitySettings.passwordPolicy.minLength}
                    onChange={(e) => setSecuritySettings({
                      ...securitySettings,
                      passwordPolicy: { ...securitySettings.passwordPolicy, minLength: parseInt(e.target.value) }
                    })}
                    min="6"
                    className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div className="flex flex-wrap gap-4">
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={securitySettings.passwordPolicy.requireUppercase}
                      onChange={(e) => setSecuritySettings({
                        ...securitySettings,
                        passwordPolicy: { ...securitySettings.passwordPolicy, requireUppercase: e.target.checked }
                      })}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">Require Uppercase</span>
                  </label>

                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={securitySettings.passwordPolicy.requireLowercase}
                      onChange={(e) => setSecuritySettings({
                        ...securitySettings,
                        passwordPolicy: { ...securitySettings.passwordPolicy, requireLowercase: e.target.checked }
                      })}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">Require Lowercase</span>
                  </label>

                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={securitySettings.passwordPolicy.requireNumbers}
                      onChange={(e) => setSecuritySettings({
                        ...securitySettings,
                        passwordPolicy: { ...securitySettings.passwordPolicy, requireNumbers: e.target.checked }
                      })}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">Require Numbers</span>
                  </label>

                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={securitySettings.passwordPolicy.requireSpecialChars}
                      onChange={(e) => setSecuritySettings({
                        ...securitySettings,
                        passwordPolicy: { ...securitySettings.passwordPolicy, requireSpecialChars: e.target.checked }
                      })}
                      className="w-4 h-4 text-blue-600 rounded"
                    />
                    <span className="text-sm text-gray-700 dark:text-gray-300">Require Special Characters</span>
                  </label>
                </div>
              </div>
            </div>

            {/* Session Timeout */}
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Session Timeout</h3>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Timeout (minutes)
                </label>
                <input
                  type="number"
                  value={securitySettings.sessionTimeout}
                  onChange={(e) => setSecuritySettings({ ...securitySettings, sessionTimeout: parseInt(e.target.value) })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>
            </div>

            {/* Two-Factor Auth */}
            <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={securitySettings.twoFactorAuth}
                  onChange={(e) => setSecuritySettings({ ...securitySettings, twoFactorAuth: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Enable Two-Factor Authentication</span>
              </label>
            </div>
          </div>
        )}

        {/* Maintenance Settings */}
        {activeTab === 'maintenance' && (
          <div className="space-y-6">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Maintenance Mode
            </h2>

            <div className="mb-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={maintenanceSettings.enabled}
                  onChange={(e) => setMaintenanceSettings({ ...maintenanceSettings, enabled: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Enable Maintenance Mode</span>
              </label>
            </div>

            {maintenanceSettings.enabled && (
              <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 rounded-lg">
                <p className="text-yellow-800 dark:text-yellow-300 text-sm">
                  Warning: Enabling maintenance mode will block access to the platform for all users except allowed IPs.
                </p>
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Maintenance Message
              </label>
              <textarea
                value={maintenanceSettings.message}
                onChange={(e) => setMaintenanceSettings({ ...maintenanceSettings, message: e.target.value })}
                rows={4}
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Allowed IPs (one per line)
              </label>
              <textarea
                value={maintenanceSettings.allowedIPs.join('\n')}
                onChange={(e) => setMaintenanceSettings({
                  ...maintenanceSettings,
                  allowedIPs: e.target.value.split('\n').filter(ip => ip.trim())
                })}
                rows={4}
                placeholder="127.0.0.1&#10;192.168.1.1"
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono text-sm"
              />
            </div>
          </div>
        )}

        {/* License Settings */}
        {activeTab === 'license' && (
          <div className="space-y-6">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              License Information
            </h2>

            <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-700 rounded-lg p-6">
              <div className="flex items-start gap-4">
                <div className="flex-shrink-0">
                  <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
                    licenseSettings.status === 'active' ? 'bg-green-500' :
                    licenseSettings.status === 'expired' ? 'bg-red-500' : 'bg-gray-500'
                  }`}>
                    <KeyIcon className="w-6 h-6 text-white" />
                  </div>
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                    License Status: <span className={`${
                      licenseSettings.status === 'active' ? 'text-green-600 dark:text-green-400' :
                      licenseSettings.status === 'expired' ? 'text-red-600 dark:text-red-400' :
                      'text-gray-600 dark:text-gray-400'
                    }`}>
                      {licenseSettings.status.toUpperCase()}
                    </span>
                  </h3>
                  <div className="space-y-2 text-sm text-gray-700 dark:text-gray-300">
                    <p><strong>License Key:</strong> {licenseSettings.key}</p>
                    <p><strong>Domain:</strong> {licenseSettings.domain}</p>
                    <p><strong>Expires:</strong> {licenseSettings.expiresAt}</p>
                    <p><strong>Max Users:</strong> {licenseSettings.maxUsers.toLocaleString()}</p>
                    <p><strong>Max Streams:</strong> {licenseSettings.maxStreams.toLocaleString()}</p>
                  </div>
                </div>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Update License Key
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="IPTV-PRO-XXXX-XXXX-XXXX-XXXX"
                  className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
                <button className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                  Verify
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Advanced Settings */}
        {activeTab === 'advanced' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                Advanced Settings
              </h2>
              <button
                onClick={handleClearCache}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Clear Cache
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Log Level
                </label>
                <select
                  value={advancedSettings.logLevel}
                  onChange={(e) => setAdvancedSettings({ ...advancedSettings, logLevel: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="debug">Debug</option>
                  <option value="info">Info</option>
                  <option value="warning">Warning</option>
                  <option value="error">Error</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Cache TTL (seconds)
                </label>
                <input
                  type="number"
                  value={advancedSettings.cacheTTL}
                  onChange={(e) => setAdvancedSettings({ ...advancedSettings, cacheTTL: parseInt(e.target.value) })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Queue Driver
                </label>
                <select
                  value={advancedSettings.queueDriver}
                  onChange={(e) => setAdvancedSettings({ ...advancedSettings, queueDriver: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="redis">Redis</option>
                  <option value="database">Database</option>
                  <option value="sync">Sync</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Session Driver
                </label>
                <select
                  value={advancedSettings.sessionDriver}
                  onChange={(e) => setAdvancedSettings({ ...advancedSettings, sessionDriver: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="redis">Redis</option>
                  <option value="database">Database</option>
                  <option value="cookie">Cookie</option>
                  <option value="file">File</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Broadcast Driver
                </label>
                <select
                  value={advancedSettings.broadcastDriver}
                  onChange={(e) => setAdvancedSettings({ ...advancedSettings, broadcastDriver: e.target.value as any })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="redis">Redis</option>
                  <option value="pusher">Pusher</option>
                  <option value="log">Log</option>
                </select>
              </div>
            </div>

            <div className="flex flex-wrap gap-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={advancedSettings.debugMode}
                  onChange={(e) => setAdvancedSettings({ ...advancedSettings, debugMode: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">Debug Mode</span>
              </label>

              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={advancedSettings.cacheEnabled}
                  onChange={(e) => setAdvancedSettings({ ...advancedSettings, cacheEnabled: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">Cache Enabled</span>
              </label>
            </div>
          </div>
        )}

        {/* Save Button */}
        <div className="mt-8 flex items-center justify-between pt-6 border-t border-gray-200 dark:border-gray-700">
          <div>
            {saved && (
              <span className="text-green-600 dark:text-green-400 text-sm">
                Settings saved successfully!
              </span>
            )}
          </div>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
          >
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default Settings;
