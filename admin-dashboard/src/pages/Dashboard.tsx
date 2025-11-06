import { useEffect, useState } from 'react';
import {
  UsersIcon,
  FilmIcon,
  CurrencyDollarIcon,
  SignalIcon,
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon,
} from '@heroicons/react/24/outline';
import { analyticsAPI } from '../services/api';
import { LineChart, Line, AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  totalStreams: number;
  activeStreams: number;
  totalRevenue: number;
  monthlyRevenue: number;
  activeSessions: number;
  bandwidth: number;
}

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats>({
    totalUsers: 0,
    activeUsers: 0,
    totalStreams: 0,
    activeStreams: 0,
    totalRevenue: 0,
    monthlyRevenue: 0,
    activeSessions: 0,
    bandwidth: 0,
  });
  const [loading, setLoading] = useState(true);

  // Mock data for charts
  const userGrowthData = [
    { name: 'Jan', users: 400, active: 240 },
    { name: 'Feb', users: 800, active: 480 },
    { name: 'Mar', users: 1200, active: 720 },
    { name: 'Apr', users: 1800, active: 1080 },
    { name: 'May', users: 2500, active: 1500 },
    { name: 'Jun', users: 3200, active: 1920 },
    { name: 'Jul', users: 4100, active: 2460 },
  ];

  const revenueData = [
    { name: 'Week 1', revenue: 4000 },
    { name: 'Week 2', revenue: 3000 },
    { name: 'Week 3', revenue: 5000 },
    { name: 'Week 4', revenue: 7800 },
  ];

  const streamTypeData = [
    { name: 'Live TV', value: 400, color: '#3b82f6' },
    { name: 'VOD', value: 300, color: '#10b981' },
    { name: 'Series', value: 200, color: '#f59e0b' },
    { name: 'Radio', value: 100, color: '#ef4444' },
  ];

  const topStreams = [
    { name: 'ESPN HD', viewers: 2345, change: 12.5 },
    { name: 'CNN', viewers: 1876, change: -3.2 },
    { name: 'BBC World', viewers: 1654, change: 8.7 },
    { name: 'Discovery', viewers: 1432, change: 15.3 },
    { name: 'HBO', viewers: 1287, change: -1.5 },
  ];

  useEffect(() => {
    loadDashboardStats();
  }, []);

  const loadDashboardStats = async () => {
    try {
      // In production, fetch from API
      // const response = await analyticsAPI.getDashboardStats();
      // setStats(response.data);

      // Mock data for now
      setTimeout(() => {
        setStats({
          totalUsers: 4158,
          activeUsers: 2495,
          totalStreams: 523,
          activeStreams: 318,
          totalRevenue: 89765,
          monthlyRevenue: 12450,
          activeSessions: 847,
          bandwidth: 12.5, // TB
        });
        setLoading(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to load dashboard stats:', error);
      setLoading(false);
    }
  };

  const StatCard = ({ title, value, change, icon: Icon, color }: any) => (
    <div className="stat-card">
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</p>
          <p className="text-3xl font-bold mt-2 text-gray-900 dark:text-white">
            {loading ? (
              <div className="h-8 w-24 skeleton rounded" />
            ) : (
              value
            )}
          </p>
          {change !== undefined && !loading && (
            <div className="flex items-center mt-2">
              {change >= 0 ? (
                <ArrowTrendingUpIcon className="h-4 w-4 text-green-500 mr-1" />
              ) : (
                <ArrowTrendingDownIcon className="h-4 w-4 text-red-500 mr-1" />
              )}
              <span className={`text-sm font-medium ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                {Math.abs(change)}%
              </span>
              <span className="text-sm text-gray-500 ml-2">vs last month</span>
            </div>
          )}
        </div>
        <div className={`stat-card-hover p-4 rounded-2xl ${color}`}>
          <Icon className="h-8 w-8 text-white" />
        </div>
      </div>
    </div>
  );

  return (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Users"
          value={stats.totalUsers.toLocaleString()}
          change={15.3}
          icon={UsersIcon}
          color="bg-gradient-to-br from-blue-500 to-blue-700"
        />
        <StatCard
          title="Active Streams"
          value={stats.activeStreams}
          change={8.2}
          icon={FilmIcon}
          color="bg-gradient-to-br from-green-500 to-green-700"
        />
        <StatCard
          title="Monthly Revenue"
          value={`$${stats.monthlyRevenue.toLocaleString()}`}
          change={12.5}
          icon={CurrencyDollarIcon}
          color="bg-gradient-to-br from-yellow-500 to-yellow-700"
        />
        <StatCard
          title="Active Sessions"
          value={stats.activeSessions}
          change={-2.1}
          icon={SignalIcon}
          color="bg-gradient-to-br from-purple-500 to-purple-700"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* User Growth Chart */}
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">User Growth</h3>
          {loading ? (
            <div className="h-80 skeleton rounded" />
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={userGrowthData}>
                <defs>
                  <linearGradient id="colorUsers" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.8}/>
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="name" stroke="#6b7280" />
                <YAxis stroke="#6b7280" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1f2937',
                    border: 'none',
                    borderRadius: '8px',
                    color: '#fff',
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="users"
                  stroke="#3b82f6"
                  fillOpacity={1}
                  fill="url(#colorUsers)"
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Revenue Chart */}
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Weekly Revenue</h3>
          {loading ? (
            <div className="h-80 skeleton rounded" />
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={revenueData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="name" stroke="#6b7280" />
                <YAxis stroke="#6b7280" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1f2937',
                    border: 'none',
                    borderRadius: '8px',
                    color: '#fff',
                  }}
                />
                <Bar dataKey="revenue" fill="#10b981" radius={[8, 8, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Stream Type Distribution */}
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Content Distribution</h3>
          {loading ? (
            <div className="h-64 skeleton rounded" />
          ) : (
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={streamTypeData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {streamTypeData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Top Streams */}
        <div className="card p-6 lg:col-span-2">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Top Streams</h3>
          {loading ? (
            <div className="space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-12 skeleton rounded" />
              ))}
            </div>
          ) : (
            <div className="space-y-3">
              {topStreams.map((stream, index) => (
                <div
                  key={stream.name}
                  className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-dark-800 transition-colors"
                >
                  <div className="flex items-center space-x-4">
                    <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 font-semibold text-sm">
                      {index + 1}
                    </div>
                    <div>
                      <p className="font-medium text-gray-900 dark:text-white">{stream.name}</p>
                      <p className="text-sm text-gray-500">{stream.viewers.toLocaleString()} viewers</p>
                    </div>
                  </div>
                  <div className={`flex items-center space-x-1 ${stream.change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                    {stream.change >= 0 ? (
                      <ArrowTrendingUpIcon className="h-4 w-4" />
                    ) : (
                      <ArrowTrendingDownIcon className="h-4 w-4" />
                    )}
                    <span className="text-sm font-medium">{Math.abs(stream.change)}%</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="card p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Quick Actions</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <button className="btn-primary py-3">Add User</button>
          <button className="btn-primary py-3">Add Stream</button>
          <button className="btn-primary py-3">Create Invoice</button>
          <button className="btn-primary py-3">View Reports</button>
        </div>
      </div>
    </div>
  );
}
