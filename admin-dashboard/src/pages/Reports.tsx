import React, { useState, useEffect } from 'react';
import {
  ChartBarIcon,
  ArrowTrendingUpIcon,
  ArrowDownTrayIcon,
  CalendarIcon,
  FunnelIcon,
  UserGroupIcon,
  CurrencyDollarIcon,
  PlayCircleIcon,
  DevicePhoneMobileIcon,
} from '@heroicons/react/24/outline';

interface ReportStats {
  revenue: {
    today: number;
    yesterday: number;
    thisMonth: number;
    lastMonth: number;
    trend: number;
  };
  users: {
    total: number;
    active: number;
    new: number;
    churn: number;
    trend: number;
  };
  streams: {
    total: number;
    live: number;
    vod: number;
    views: number;
    trend: number;
  };
  devices: {
    total: number;
    active: number;
    blocked: number;
    sessions: number;
  };
}

interface ChartData {
  label: string;
  value: number;
  color?: string;
}

interface TimeRange {
  value: string;
  label: string;
}

const Reports: React.FC = () => {
  const [stats, setStats] = useState<ReportStats>({
    revenue: { today: 0, yesterday: 0, thisMonth: 0, lastMonth: 0, trend: 0 },
    users: { total: 0, active: 0, new: 0, churn: 0, trend: 0 },
    streams: { total: 0, live: 0, vod: 0, views: 0, trend: 0 },
    devices: { total: 0, active: 0, blocked: 0, sessions: 0 },
  });

  const [timeRange, setTimeRange] = useState<string>('30d');
  const [reportType, setReportType] = useState<string>('revenue');
  const [revenueData, setRevenueData] = useState<ChartData[]>([]);
  const [userGrowthData, setUserGrowthData] = useState<ChartData[]>([]);
  const [topStreamsData, setTopStreamsData] = useState<ChartData[]>([]);
  const [resellerPerformance, setResellerPerformance] = useState<ChartData[]>([]);

  const timeRanges: TimeRange[] = [
    { value: '7d', label: 'Last 7 Days' },
    { value: '30d', label: 'Last 30 Days' },
    { value: '90d', label: 'Last 90 Days' },
    { value: '12m', label: 'Last 12 Months' },
    { value: 'ytd', label: 'Year to Date' },
    { value: 'custom', label: 'Custom Range' },
  ];

  useEffect(() => {
    loadMockData();
  }, [timeRange]);

  const loadMockData = () => {
    // Mock statistics
    const mockStats: ReportStats = {
      revenue: {
        today: 3450.50,
        yesterday: 3120.00,
        thisMonth: 89234.75,
        lastMonth: 82150.20,
        trend: 8.6,
      },
      users: {
        total: 12890,
        active: 9567,
        new: 456,
        churn: 89,
        trend: 12.3,
      },
      streams: {
        total: 3542,
        live: 458,
        vod: 3084,
        views: 456789,
        trend: 15.7,
      },
      devices: {
        total: 18234,
        active: 14567,
        blocked: 234,
        sessions: 8934,
      },
    };

    // Mock revenue chart data (last 30 days)
    const mockRevenue: ChartData[] = [];
    for (let i = 29; i >= 0; i--) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      mockRevenue.push({
        label: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        value: Math.floor(Math.random() * 5000) + 2000,
        color: '#3B82F6',
      });
    }

    // Mock user growth data
    const mockUserGrowth: ChartData[] = [];
    for (let i = 11; i >= 0; i--) {
      const date = new Date();
      date.setMonth(date.getMonth() - i);
      mockUserGrowth.push({
        label: date.toLocaleDateString('en-US', { month: 'short', year: '2-digit' }),
        value: Math.floor(Math.random() * 500) + 300,
        color: '#10B981',
      });
    }

    // Mock top streams
    const mockTopStreams: ChartData[] = [
      { label: 'CNN International', value: 45230, color: '#EF4444' },
      { label: 'ESPN Sports HD', value: 38920, color: '#F59E0B' },
      { label: 'BBC World News', value: 32450, color: '#3B82F6' },
      { label: 'National Geographic', value: 28760, color: '#10B981' },
      { label: 'Discovery Channel', value: 21340, color: '#8B5CF6' },
      { label: 'The Dark Knight', value: 18650, color: '#EC4899' },
      { label: 'Inception', value: 15890, color: '#6366F1' },
      { label: 'Shawshank Redemption', value: 14230, color: '#14B8A6' },
    ];

    // Mock reseller performance
    const mockResellers: ChartData[] = [
      { label: 'John Reseller', value: 125, color: '#3B82F6' },
      { label: 'Sarah Distributor', value: 98, color: '#10B981' },
      { label: 'Mike Channel', value: 87, color: '#F59E0B' },
      { label: 'Lisa Network', value: 76, color: '#EF4444' },
      { label: 'Tom Stream', value: 65, color: '#8B5CF6' },
    ];

    setStats(mockStats);
    setRevenueData(mockRevenue);
    setUserGrowthData(mockUserGrowth);
    setTopStreamsData(mockTopStreams);
    setResellerPerformance(mockResellers);
  };

  const exportReport = (format: 'csv' | 'pdf' | 'excel') => {
    console.log(`Exporting report as ${format}...`);
    alert(`Report export as ${format.toUpperCase()} will be implemented with backend integration.`);
  };

  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
    }).format(amount);
  };

  const formatNumber = (num: number): string => {
    return new Intl.NumberFormat('en-US').format(num);
  };

  const getTrendColor = (trend: number): string => {
    return trend >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400';
  };

  const getTrendIcon = (trend: number) => {
    return trend >= 0 ? '↑' : '↓';
  };

  const renderBarChart = (data: ChartData[], maxHeight: number = 200) => {
    const maxValue = Math.max(...data.map(d => d.value));

    return (
      <div className="flex items-end justify-between gap-2 h-64 px-4">
        {data.slice(0, 15).map((item, index) => {
          const height = (item.value / maxValue) * maxHeight;
          return (
            <div key={index} className="flex-1 flex flex-col items-center gap-2">
              <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-t-lg relative group" style={{ height: `${height}px`, minHeight: '10px', backgroundColor: item.color || '#3B82F6' }}>
                <div className="absolute -top-8 left-1/2 transform -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity bg-gray-900 text-white text-xs px-2 py-1 rounded whitespace-nowrap">
                  {formatNumber(item.value)}
                </div>
              </div>
              <span className="text-xs text-gray-600 dark:text-gray-400 transform rotate-45 origin-left whitespace-nowrap">
                {item.label}
              </span>
            </div>
          );
        })}
      </div>
    );
  };

  const renderHorizontalBarChart = (data: ChartData[]) => {
    const maxValue = Math.max(...data.map(d => d.value));

    return (
      <div className="space-y-4 px-4">
        {data.map((item, index) => {
          const width = (item.value / maxValue) * 100;
          return (
            <div key={index} className="space-y-1">
              <div className="flex justify-between text-sm">
                <span className="text-gray-700 dark:text-gray-300 font-medium">{item.label}</span>
                <span className="text-gray-600 dark:text-gray-400">{formatNumber(item.value)}</span>
              </div>
              <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-4">
                <div
                  className="h-4 rounded-full transition-all duration-500"
                  style={{ width: `${width}%`, backgroundColor: item.color || '#3B82F6' }}
                />
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="p-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
      {/* Header */}
      <div className="mb-6 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <ChartBarIcon className="w-8 h-8 text-blue-500" />
            Reports & Analytics
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Comprehensive insights and performance metrics
          </p>
        </div>

        <div className="flex flex-wrap gap-3">
          {/* Time Range Selector */}
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          >
            {timeRanges.map(range => (
              <option key={range.value} value={range.value}>{range.label}</option>
            ))}
          </select>

          {/* Export Buttons */}
          <div className="flex gap-2">
            <button
              onClick={() => exportReport('csv')}
              className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg flex items-center gap-2 transition-colors"
            >
              <ArrowDownTrayIcon className="w-5 h-5" />
              CSV
            </button>
            <button
              onClick={() => exportReport('pdf')}
              className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg flex items-center gap-2 transition-colors"
            >
              <ArrowDownTrayIcon className="w-5 h-5" />
              PDF
            </button>
            <button
              onClick={() => exportReport('excel')}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg flex items-center gap-2 transition-colors"
            >
              <ArrowDownTrayIcon className="w-5 h-5" />
              Excel
            </button>
          </div>
        </div>
      </div>

      {/* Key Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {/* Revenue Card */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">Revenue (Month)</h3>
            <CurrencyDollarIcon className="w-8 h-8 text-green-500" />
          </div>
          <div className="space-y-2">
            <p className="text-3xl font-bold text-gray-900 dark:text-white">
              {formatCurrency(stats.revenue.thisMonth)}
            </p>
            <div className="flex items-center gap-2">
              <span className={`text-sm font-medium ${getTrendColor(stats.revenue.trend)}`}>
                {getTrendIcon(stats.revenue.trend)} {Math.abs(stats.revenue.trend)}%
              </span>
              <span className="text-xs text-gray-500">vs last month</span>
            </div>
            <div className="pt-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-600 dark:text-gray-400">
              Today: {formatCurrency(stats.revenue.today)}
            </div>
          </div>
        </div>

        {/* Users Card */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Users</h3>
            <UserGroupIcon className="w-8 h-8 text-blue-500" />
          </div>
          <div className="space-y-2">
            <p className="text-3xl font-bold text-gray-900 dark:text-white">
              {formatNumber(stats.users.total)}
            </p>
            <div className="flex items-center gap-2">
              <span className={`text-sm font-medium ${getTrendColor(stats.users.trend)}`}>
                {getTrendIcon(stats.users.trend)} {Math.abs(stats.users.trend)}%
              </span>
              <span className="text-xs text-gray-500">growth rate</span>
            </div>
            <div className="pt-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-600 dark:text-gray-400">
              Active: {formatNumber(stats.users.active)} ({((stats.users.active / stats.users.total) * 100).toFixed(1)}%)
            </div>
          </div>
        </div>

        {/* Streams Card */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Streams</h3>
            <PlayCircleIcon className="w-8 h-8 text-purple-500" />
          </div>
          <div className="space-y-2">
            <p className="text-3xl font-bold text-gray-900 dark:text-white">
              {formatNumber(stats.streams.total)}
            </p>
            <div className="flex items-center gap-2">
              <span className={`text-sm font-medium ${getTrendColor(stats.streams.trend)}`}>
                {getTrendIcon(stats.streams.trend)} {Math.abs(stats.streams.trend)}%
              </span>
              <span className="text-xs text-gray-500">views increase</span>
            </div>
            <div className="pt-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-600 dark:text-gray-400">
              Live: {formatNumber(stats.streams.live)} | VOD: {formatNumber(stats.streams.vod)}
            </div>
          </div>
        </div>

        {/* Devices Card */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">Active Devices</h3>
            <DevicePhoneMobileIcon className="w-8 h-8 text-orange-500" />
          </div>
          <div className="space-y-2">
            <p className="text-3xl font-bold text-gray-900 dark:text-white">
              {formatNumber(stats.devices.active)}
            </p>
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                {formatNumber(stats.devices.sessions)} sessions
              </span>
            </div>
            <div className="pt-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-600 dark:text-gray-400">
              Total: {formatNumber(stats.devices.total)} | Blocked: {formatNumber(stats.devices.blocked)}
            </div>
          </div>
        </div>
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Revenue Chart */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Revenue Trend</h3>
            <select className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              <option>Daily</option>
              <option>Weekly</option>
              <option>Monthly</option>
            </select>
          </div>
          {renderBarChart(revenueData)}
          <div className="mt-4 text-center text-sm text-gray-600 dark:text-gray-400">
            Last 30 Days Revenue Performance
          </div>
        </div>

        {/* User Growth Chart */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">User Growth</h3>
            <div className="flex gap-2">
              <span className="px-2 py-1 text-xs bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 rounded">
                New Users
              </span>
            </div>
          </div>
          {renderBarChart(userGrowthData)}
          <div className="mt-4 text-center text-sm text-gray-600 dark:text-gray-400">
            Monthly User Acquisition Rate
          </div>
        </div>
      </div>

      {/* Top Content & Reseller Performance */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Streams */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Top Performing Streams</h3>
            <span className="text-sm text-gray-500 dark:text-gray-400">By Views</span>
          </div>
          {renderHorizontalBarChart(topStreamsData)}
          <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-center">
            <button className="text-blue-600 dark:text-blue-400 hover:underline text-sm">
              View Full Report →
            </button>
          </div>
        </div>

        {/* Reseller Performance */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Reseller Performance</h3>
            <span className="text-sm text-gray-500 dark:text-gray-400">By Customers</span>
          </div>
          {renderHorizontalBarChart(resellerPerformance)}
          <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-center">
            <button className="text-blue-600 dark:text-blue-400 hover:underline text-sm">
              View All Resellers →
            </button>
          </div>
        </div>
      </div>

      {/* Additional Insights */}
      <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg shadow p-6 text-white">
          <h4 className="text-lg font-semibold mb-2">Average Revenue per User</h4>
          <p className="text-3xl font-bold">{formatCurrency(stats.revenue.thisMonth / stats.users.active)}</p>
          <p className="text-sm opacity-90 mt-2">↑ 5.2% from last month</p>
        </div>

        <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-lg shadow p-6 text-white">
          <h4 className="text-lg font-semibold mb-2">Customer Retention Rate</h4>
          <p className="text-3xl font-bold">94.3%</p>
          <p className="text-sm opacity-90 mt-2">↑ 2.1% improvement</p>
        </div>

        <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg shadow p-6 text-white">
          <h4 className="text-lg font-semibold mb-2">Content Engagement</h4>
          <p className="text-3xl font-bold">68.5%</p>
          <p className="text-sm opacity-90 mt-2">Average watch time</p>
        </div>
      </div>
    </div>
  );
};

export default Reports;
