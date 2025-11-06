import React, { useState, useEffect } from 'react';
import {
  FilmIcon,
  PlayIcon,
  PauseIcon,
  StopIcon,
  ArrowPathIcon,
  PlusIcon,
  ChartBarIcon,
  CpuChipIcon,
} from '@heroicons/react/24/outline';

// Interfaces
interface TranscodingJob {
  id: number;
  job_name: string;
  job_type: 'live_transcode' | 'vod_transcode' | 'recording' | 'clip_generation';
  status: 'pending' | 'queued' | 'processing' | 'completed' | 'failed' | 'cancelled';
  priority: number;
  stream_id?: number;
  source_url: string;
  target_url: string;
  target_quality: 'sd' | 'hd' | 'fhd' | 'uhd';
  target_resolution: string;
  target_bitrate: number;
  encoder: string;
  audio_codec: string;
  audio_bitrate: number;
  progress: number;
  current_time: number;
  fps: number;
  speed: number;
  bitrate?: string;
  worker_id?: string;
  started_at?: string;
  completed_at?: string;
  failed_at?: string;
  error_message?: string;
  retry_count: number;
  max_retries: number;
  created_at: string;
  updated_at: string;
}

interface QueueStatus {
  total_jobs: number;
  pending_jobs: number;
  queued_jobs: number;
  processing_jobs: number;
  completed_jobs: number;
  failed_jobs: number;
  active_workers: number;
  queue_length: number;
  avg_wait_time: number;
}

interface Worker {
  id: string;
  hostname: string;
  status: 'active' | 'idle' | 'offline';
  current_jobs: number;
  max_jobs: number;
  cpu_usage: number;
  memory_usage: number;
  last_heartbeat: string;
}

interface TranscodingStats {
  period: string;
  total_jobs: number;
  completed_jobs: number;
  failed_jobs: number;
  cancelled_jobs: number;
  avg_processing_time: number;
  total_processing_time: number;
  avg_speed: number;
  total_input_size: number;
  total_output_size: number;
  compression_ratio: number;
  sd_jobs: number;
  hd_jobs: number;
  fhd_jobs: number;
  uhd_jobs: number;
}

const Transcoding: React.FC = () => {
  const [jobs, setJobs] = useState<TranscodingJob[]>([]);
  const [queueStatus, setQueueStatus] = useState<QueueStatus | null>(null);
  const [workers, setWorkers] = useState<Worker[]>([]);
  const [stats, setStats] = useState<TranscodingStats | null>(null);
  const [selectedTab, setSelectedTab] = useState<'jobs' | 'queue' | 'workers' | 'stats'>('jobs');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

  // Load mock data
  useEffect(() => {
    loadMockData();
  }, []);

  // Auto-refresh every 5 seconds
  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(() => {
        loadMockData();
        setLastRefresh(new Date());
      }, 5000);

      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  const loadMockData = () => {
    // Mock jobs
    const mockJobs: TranscodingJob[] = [
      {
        id: 1,
        job_name: 'Movie_1080p_Transcode',
        job_type: 'vod_transcode',
        status: 'completed',
        priority: 3,
        stream_id: 1,
        source_url: 'https://cdn.example.com/source/movie1.mkv',
        target_url: 'https://cdn.example.com/output/movie1_1080p.mp4',
        target_quality: 'fhd',
        target_resolution: '1920x1080',
        target_bitrate: 5000,
        encoder: 'libx264',
        audio_codec: 'aac',
        audio_bitrate: 192,
        progress: 100,
        current_time: 7200,
        fps: 30,
        speed: 1.2,
        bitrate: '5000k',
        worker_id: 'worker-01',
        started_at: new Date(Date.now() - 7200000).toISOString(),
        completed_at: new Date(Date.now() - 1800000).toISOString(),
        retry_count: 0,
        max_retries: 3,
        created_at: new Date(Date.now() - 7200000).toISOString(),
        updated_at: new Date(Date.now() - 1800000).toISOString(),
      },
      {
        id: 2,
        job_name: 'Live_Sports_Stream',
        job_type: 'live_transcode',
        status: 'processing',
        priority: 1,
        stream_id: 5,
        source_url: 'rtmp://source.example.com/live/sports',
        target_url: 'https://cdn.example.com/live/sports/index.m3u8',
        target_quality: 'fhd',
        target_resolution: '1920x1080',
        target_bitrate: 6000,
        encoder: 'h264_nvenc',
        audio_codec: 'aac',
        audio_bitrate: 128,
        progress: 45.5,
        current_time: 2730,
        fps: 60,
        speed: 1.05,
        bitrate: '6000k',
        worker_id: 'worker-02',
        started_at: new Date(Date.now() - 1800000).toISOString(),
        retry_count: 0,
        max_retries: 3,
        created_at: new Date(Date.now() - 1800000).toISOString(),
        updated_at: new Date().toISOString(),
      },
      {
        id: 3,
        job_name: 'Documentary_4K',
        job_type: 'vod_transcode',
        status: 'queued',
        priority: 7,
        source_url: 'https://cdn.example.com/source/documentary_4k.mov',
        target_url: 'https://cdn.example.com/output/documentary_4k.mp4',
        target_quality: 'uhd',
        target_resolution: '3840x2160',
        target_bitrate: 15000,
        encoder: 'libx265',
        audio_codec: 'aac',
        audio_bitrate: 256,
        progress: 0,
        current_time: 0,
        fps: 0,
        speed: 0,
        retry_count: 0,
        max_retries: 3,
        created_at: new Date(Date.now() - 900000).toISOString(),
        updated_at: new Date(Date.now() - 900000).toISOString(),
      },
      {
        id: 4,
        job_name: 'Series_S01E01_HD',
        job_type: 'vod_transcode',
        status: 'completed',
        priority: 5,
        stream_id: 2,
        source_url: 'https://cdn.example.com/source/series_s01e01.avi',
        target_url: 'https://cdn.example.com/output/series_s01e01_720p.mp4',
        target_quality: 'hd',
        target_resolution: '1280x720',
        target_bitrate: 2500,
        encoder: 'libx264',
        audio_codec: 'aac',
        audio_bitrate: 128,
        progress: 100,
        current_time: 2700,
        fps: 24,
        speed: 1.15,
        worker_id: 'worker-01',
        started_at: new Date(Date.now() - 3600000).toISOString(),
        completed_at: new Date(Date.now() - 2100000).toISOString(),
        retry_count: 0,
        max_retries: 3,
        created_at: new Date(Date.now() - 3600000).toISOString(),
        updated_at: new Date(Date.now() - 2100000).toISOString(),
      },
      {
        id: 5,
        job_name: 'Recording_CH101',
        job_type: 'recording',
        status: 'failed',
        priority: 5,
        stream_id: 8,
        source_url: 'rtmp://source.example.com/live/ch101',
        target_url: 'https://cdn.example.com/recordings/ch101_20251106.mp4',
        target_quality: 'hd',
        target_resolution: '1280x720',
        target_bitrate: 2500,
        encoder: 'libx264',
        audio_codec: 'aac',
        audio_bitrate: 128,
        progress: 12.3,
        current_time: 369,
        fps: 30,
        speed: 0.8,
        worker_id: 'worker-03',
        started_at: new Date(Date.now() - 2700000).toISOString(),
        failed_at: new Date(Date.now() - 1500000).toISOString(),
        error_message: 'Source stream disconnected unexpectedly',
        retry_count: 1,
        max_retries: 3,
        created_at: new Date(Date.now() - 2700000).toISOString(),
        updated_at: new Date(Date.now() - 1500000).toISOString(),
      },
      {
        id: 6,
        job_name: 'Clip_Highlights',
        job_type: 'clip_generation',
        status: 'pending',
        priority: 4,
        stream_id: 3,
        source_url: 'https://cdn.example.com/source/full_match.mp4',
        target_url: 'https://cdn.example.com/clips/highlights.mp4',
        target_quality: 'hd',
        target_resolution: '1280x720',
        target_bitrate: 3000,
        encoder: 'libx264',
        audio_codec: 'aac',
        audio_bitrate: 128,
        progress: 0,
        current_time: 0,
        fps: 0,
        speed: 0,
        retry_count: 0,
        max_retries: 3,
        created_at: new Date(Date.now() - 300000).toISOString(),
        updated_at: new Date(Date.now() - 300000).toISOString(),
      },
    ];

    setJobs(mockJobs);

    // Mock queue status
    setQueueStatus({
      total_jobs: 6,
      pending_jobs: 1,
      queued_jobs: 1,
      processing_jobs: 1,
      completed_jobs: 2,
      failed_jobs: 1,
      active_workers: 2,
      queue_length: 2,
      avg_wait_time: 45,
    });

    // Mock workers
    setWorkers([
      {
        id: 'worker-01',
        hostname: 'transcode-worker-01.example.com',
        status: 'active',
        current_jobs: 0,
        max_jobs: 2,
        cpu_usage: 45.5,
        memory_usage: 62.3,
        last_heartbeat: new Date().toISOString(),
      },
      {
        id: 'worker-02',
        hostname: 'transcode-worker-02.example.com',
        status: 'active',
        current_jobs: 1,
        max_jobs: 2,
        cpu_usage: 87.2,
        memory_usage: 78.9,
        last_heartbeat: new Date().toISOString(),
      },
      {
        id: 'worker-03',
        hostname: 'transcode-worker-03.example.com',
        status: 'idle',
        current_jobs: 0,
        max_jobs: 2,
        cpu_usage: 12.1,
        memory_usage: 34.5,
        last_heartbeat: new Date(Date.now() - 120000).toISOString(),
      },
    ]);

    // Mock stats
    setStats({
      period: 'Last 7 days',
      total_jobs: 125,
      completed_jobs: 102,
      failed_jobs: 18,
      cancelled_jobs: 5,
      avg_processing_time: 1840,
      total_processing_time: 229800,
      avg_speed: 1.12,
      total_input_size: 524288000000,
      total_output_size: 314572800000,
      compression_ratio: 1.67,
      sd_jobs: 15,
      hd_jobs: 58,
      fhd_jobs: 38,
      uhd_jobs: 14,
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
      case 'processing': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
      case 'queued': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300';
      case 'pending': return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
      case 'failed': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
      case 'cancelled': return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300';
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
  };

  const getQualityBadge = (quality: string) => {
    const colors: Record<string, string> = {
      sd: 'bg-gray-200 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
      hd: 'bg-blue-200 text-blue-800 dark:bg-blue-700 dark:text-blue-300',
      fhd: 'bg-purple-200 text-purple-800 dark:bg-purple-700 dark:text-purple-300',
      uhd: 'bg-pink-200 text-pink-800 dark:bg-pink-700 dark:text-pink-300',
    };
    return colors[quality] || colors.sd;
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m ${secs}s`;
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const cancelJob = (jobId: number) => {
    if (confirm(`Are you sure you want to cancel job #${jobId}?`)) {
      alert(`Job #${jobId} has been cancelled`);
      loadMockData();
    }
  };

  const retryJob = (jobId: number) => {
    alert(`Job #${jobId} has been queued for retry`);
    loadMockData();
  };

  const deleteJob = (jobId: number) => {
    if (confirm(`Are you sure you want to delete job #${jobId}?`)) {
      alert(`Job #${jobId} has been deleted`);
      loadMockData();
    }
  };

  const filteredJobs = filterStatus === 'all' ? jobs : jobs.filter(job => job.status === filterStatus);

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            FFmpeg Transcoding
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage video transcoding jobs and monitor processing queue
          </p>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded"
            />
            Auto-refresh
          </label>
          <button
            onClick={loadMockData}
            className="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2"
          >
            <ArrowPathIcon className="w-5 h-5" />
            Refresh
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
          >
            <PlusIcon className="w-5 h-5" />
            New Job
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="mb-6 border-b border-gray-200 dark:border-gray-700">
        <div className="flex gap-4 -mb-px">
          {['jobs', 'queue', 'workers', 'stats'].map((tab) => (
            <button
              key={tab}
              onClick={() => setSelectedTab(tab as any)}
              className={`px-4 py-2 border-b-2 font-medium text-sm capitalize transition-colors ${
                selectedTab === tab
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
              }`}
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      {/* Jobs Tab */}
      {selectedTab === 'jobs' && (
        <div>
          {/* Filters */}
          <div className="mb-4 flex items-center gap-4">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Filter:</label>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="all">All Status</option>
              <option value="pending">Pending</option>
              <option value="queued">Queued</option>
              <option value="processing">Processing</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="cancelled">Cancelled</option>
            </select>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              Showing {filteredJobs.length} of {jobs.length} jobs
            </span>
          </div>

          {/* Jobs Table */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-900">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Job Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Quality</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Progress</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Speed</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Priority</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  {filteredJobs.map((job) => (
                    <tr key={job.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                        #{job.id}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {job.job_name}
                        </div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">
                          {job.encoder} • {job.audio_codec} • {job.target_resolution}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400 capitalize">
                        {job.job_type.replace('_', ' ')}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusColor(job.status)}`}>
                          {job.status}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2 py-1 text-xs font-semibold rounded uppercase ${getQualityBadge(job.target_quality)}`}>
                          {job.target_quality}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-2">
                          <div className="flex-1 w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                            <div
                              className={`h-2 rounded-full ${
                                job.status === 'completed' ? 'bg-green-500' :
                                job.status === 'failed' ? 'bg-red-500' :
                                job.status === 'processing' ? 'bg-blue-500' :
                                'bg-gray-400'
                              }`}
                              style={{ width: `${job.progress}%` }}
                            />
                          </div>
                          <span className="text-sm text-gray-700 dark:text-gray-300 font-medium">
                            {job.progress.toFixed(1)}%
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        {job.speed > 0 ? `${job.speed.toFixed(2)}x` : '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                        P{job.priority}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <div className="flex items-center gap-2">
                          {job.status === 'processing' && (
                            <button
                              onClick={() => cancelJob(job.id)}
                              className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                              title="Cancel"
                            >
                              <StopIcon className="w-5 h-5" />
                            </button>
                          )}
                          {job.status === 'failed' && (
                            <button
                              onClick={() => retryJob(job.id)}
                              className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                              title="Retry"
                            >
                              <ArrowPathIcon className="w-5 h-5" />
                            </button>
                          )}
                          {(job.status === 'completed' || job.status === 'failed') && (
                            <button
                              onClick={() => deleteJob(job.id)}
                              className="text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-300"
                              title="Delete"
                            >
                              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                              </svg>
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Queue Tab */}
      {selectedTab === 'queue' && queueStatus && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Total Jobs</h3>
              <FilmIcon className="w-8 h-8 text-blue-500" />
            </div>
            <p className="text-3xl font-bold text-gray-900 dark:text-white">{queueStatus.total_jobs}</p>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">Today</p>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Queue Length</h3>
              <PlayIcon className="w-8 h-8 text-yellow-500" />
            </div>
            <p className="text-3xl font-bold text-gray-900 dark:text-white">{queueStatus.queue_length}</p>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
              {queueStatus.pending_jobs} pending, {queueStatus.queued_jobs} queued
            </p>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Processing</h3>
              <CpuChipIcon className="w-8 h-8 text-green-500" />
            </div>
            <p className="text-3xl font-bold text-gray-900 dark:text-white">{queueStatus.processing_jobs}</p>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
              {queueStatus.active_workers} workers active
            </p>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Avg Wait Time</h3>
              <ChartBarIcon className="w-8 h-8 text-purple-500" />
            </div>
            <p className="text-3xl font-bold text-gray-900 dark:text-white">{queueStatus.avg_wait_time}s</p>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
              {queueStatus.completed_jobs} completed, {queueStatus.failed_jobs} failed
            </p>
          </div>
        </div>
      )}

      {/* Workers Tab */}
      {selectedTab === 'workers' && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-900">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Worker ID</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Hostname</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Jobs</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">CPU Usage</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Memory Usage</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Last Heartbeat</th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {workers.map((worker) => (
                  <tr key={worker.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                      {worker.id}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {worker.hostname}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        worker.status === 'active' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300' :
                        worker.status === 'idle' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300' :
                        'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300'
                      }`}>
                        {worker.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {worker.current_jobs} / {worker.max_jobs}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <div className="flex-1 w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                          <div
                            className={`h-2 rounded-full ${
                              worker.cpu_usage > 80 ? 'bg-red-500' :
                              worker.cpu_usage > 60 ? 'bg-yellow-500' :
                              'bg-green-500'
                            }`}
                            style={{ width: `${worker.cpu_usage}%` }}
                          />
                        </div>
                        <span className="text-sm text-gray-700 dark:text-gray-300">
                          {worker.cpu_usage.toFixed(1)}%
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <div className="flex-1 w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                          <div
                            className={`h-2 rounded-full ${
                              worker.memory_usage > 80 ? 'bg-red-500' :
                              worker.memory_usage > 60 ? 'bg-yellow-500' :
                              'bg-green-500'
                            }`}
                            style={{ width: `${worker.memory_usage}%` }}
                          />
                        </div>
                        <span className="text-sm text-gray-700 dark:text-gray-300">
                          {worker.memory_usage.toFixed(1)}%
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {new Date(worker.last_heartbeat).toLocaleTimeString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Stats Tab */}
      {selectedTab === 'stats' && stats && (
        <div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
              <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Total Jobs</h3>
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{stats.total_jobs}</p>
              <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
                <div>Completed: {stats.completed_jobs}</div>
                <div>Failed: {stats.failed_jobs}</div>
                <div>Cancelled: {stats.cancelled_jobs}</div>
              </div>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
              <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Avg Processing Time</h3>
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{formatDuration(stats.avg_processing_time)}</p>
              <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
                Total: {formatDuration(stats.total_processing_time)}
              </div>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
              <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Storage</h3>
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{formatFileSize(stats.total_output_size)}</p>
              <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
                Input: {formatFileSize(stats.total_input_size)}
              </div>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
              <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Avg Speed</h3>
              <p className="text-3xl font-bold text-gray-900 dark:text-white">{stats.avg_speed.toFixed(2)}x</p>
              <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
                Compression: {stats.compression_ratio.toFixed(2)}x
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Jobs by Quality</h3>
            <div className="grid grid-cols-4 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.sd_jobs}</div>
                <div className="text-sm text-gray-600 dark:text-gray-400">SD (360p)</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.hd_jobs}</div>
                <div className="text-sm text-gray-600 dark:text-gray-400">HD (720p)</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.fhd_jobs}</div>
                <div className="text-sm text-gray-600 dark:text-gray-400">Full HD (1080p)</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.uhd_jobs}</div>
                <div className="text-sm text-gray-600 dark:text-gray-400">4K (2160p)</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Transcoding;
