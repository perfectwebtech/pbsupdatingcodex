import React, { useState, useEffect } from 'react';
import {
  PlayCircleIcon,
  FilmIcon,
  TvIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  StarIcon,
  TrashIcon,
  PencilIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import { StarIcon as StarIconSolid } from '@heroicons/react/24/solid';

type StreamType = 'live' | 'vod';

interface Stream {
  id: number;
  name: string;
  stream_type: StreamType;
  category_id: number;
  stream_url: string;
  thumbnail?: string;
  description?: string;
  language?: string;
  country?: string;
  genre?: string;
  rating?: number;
  is_active: boolean;
  is_featured: boolean;
  view_count: number;
  duration?: number; // For VOD in seconds
  release_year?: number; // For VOD
  quality?: string; // HD, FHD, 4K
  created_at: string;
  updated_at: string;
}

interface StreamStats {
  total_streams: number;
  live_streams: number;
  vod_streams: number;
  featured_streams: number;
  active_streams: number;
  total_views: number;
}

interface StreamFormData {
  name: string;
  stream_type: StreamType;
  category_id: number;
  stream_url: string;
  thumbnail?: string;
  description?: string;
  language?: string;
  country?: string;
  genre?: string;
  rating?: number;
  is_active: boolean;
  is_featured: boolean;
  duration?: number;
  release_year?: number;
  quality?: string;
}

const Streams: React.FC = () => {
  const [streams, setStreams] = useState<Stream[]>([]);
  const [stats, setStats] = useState<StreamStats>({
    total_streams: 0,
    live_streams: 0,
    vod_streams: 0,
    featured_streams: 0,
    active_streams: 0,
    total_views: 0,
  });

  const [viewMode, setViewMode] = useState<'live' | 'vod'>('live');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [selectedStream, setSelectedStream] = useState<Stream | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterCategory, setFilterCategory] = useState<number | 'all'>('all');
  const [filterQuality, setFilterQuality] = useState<string | 'all'>('all');

  const [formData, setFormData] = useState<StreamFormData>({
    name: '',
    stream_type: 'live',
    category_id: 1,
    stream_url: '',
    thumbnail: '',
    description: '',
    language: 'English',
    country: 'USA',
    genre: '',
    rating: 0,
    is_active: true,
    is_featured: false,
    duration: 0,
    release_year: new Date().getFullYear(),
    quality: 'FHD',
  });

  // Mock data for demonstration
  useEffect(() => {
    loadMockData();
  }, []);

  const loadMockData = () => {
    const mockStreams: Stream[] = [
      // Live Streams
      {
        id: 1,
        name: 'CNN International',
        stream_type: 'live',
        category_id: 1,
        stream_url: 'https://cnn-stream.example.com/live',
        thumbnail: 'https://via.placeholder.com/400x225/1a1a2e/16a085?text=CNN',
        description: '24/7 breaking news and world coverage',
        language: 'English',
        country: 'USA',
        genre: 'News',
        is_active: true,
        is_featured: true,
        view_count: 45230,
        quality: 'FHD',
        created_at: new Date(Date.now() - 90 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 1 * 86400000).toISOString(),
      },
      {
        id: 2,
        name: 'BBC World News',
        stream_type: 'live',
        category_id: 1,
        stream_url: 'https://bbc-stream.example.com/live',
        thumbnail: 'https://via.placeholder.com/400x225/2c3e50/e74c3c?text=BBC',
        description: 'International news and current affairs',
        language: 'English',
        country: 'UK',
        genre: 'News',
        is_active: true,
        is_featured: true,
        view_count: 38920,
        quality: 'FHD',
        created_at: new Date(Date.now() - 85 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 2 * 86400000).toISOString(),
      },
      {
        id: 3,
        name: 'ESPN Sports HD',
        stream_type: 'live',
        category_id: 2,
        stream_url: 'https://espn-stream.example.com/live',
        thumbnail: 'https://via.placeholder.com/400x225/16213e/e74c3c?text=ESPN',
        description: 'Live sports coverage and highlights',
        language: 'English',
        country: 'USA',
        genre: 'Sports',
        is_active: true,
        is_featured: false,
        view_count: 62450,
        quality: 'FHD',
        created_at: new Date(Date.now() - 120 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 1 * 86400000).toISOString(),
      },
      {
        id: 4,
        name: 'National Geographic',
        stream_type: 'live',
        category_id: 3,
        stream_url: 'https://natgeo-stream.example.com/live',
        thumbnail: 'https://via.placeholder.com/400x225/f39c12/2c3e50?text=NatGeo',
        description: 'Wildlife and nature documentaries',
        language: 'English',
        country: 'USA',
        genre: 'Documentary',
        is_active: true,
        is_featured: true,
        view_count: 28760,
        quality: '4K',
        created_at: new Date(Date.now() - 100 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 3 * 86400000).toISOString(),
      },
      {
        id: 5,
        name: 'Discovery Channel',
        stream_type: 'live',
        category_id: 3,
        stream_url: 'https://discovery-stream.example.com/live',
        thumbnail: 'https://via.placeholder.com/400x225/34495e/ecf0f1?text=Discovery',
        description: 'Science, technology, and exploration',
        language: 'English',
        country: 'USA',
        genre: 'Documentary',
        is_active: true,
        is_featured: false,
        view_count: 21340,
        quality: 'FHD',
        created_at: new Date(Date.now() - 110 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 4 * 86400000).toISOString(),
      },

      // VOD Streams
      {
        id: 101,
        name: 'The Shawshank Redemption',
        stream_type: 'vod',
        category_id: 4,
        stream_url: 'https://vod.example.com/shawshank.mp4',
        thumbnail: 'https://via.placeholder.com/400x600/1a1a2e/16a085?text=Shawshank',
        description: 'Two imprisoned men bond over years, finding redemption through acts of common decency.',
        language: 'English',
        country: 'USA',
        genre: 'Drama',
        rating: 9.3,
        is_active: true,
        is_featured: true,
        view_count: 125600,
        duration: 8520, // 142 minutes
        release_year: 1994,
        quality: '4K',
        created_at: new Date(Date.now() - 200 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 5 * 86400000).toISOString(),
      },
      {
        id: 102,
        name: 'The Dark Knight',
        stream_type: 'vod',
        category_id: 4,
        stream_url: 'https://vod.example.com/dark-knight.mp4',
        thumbnail: 'https://via.placeholder.com/400x600/0f0f0f/ffffff?text=Dark+Knight',
        description: 'Batman faces the Joker, a criminal mastermind who wants to plunge Gotham into anarchy.',
        language: 'English',
        country: 'USA',
        genre: 'Action',
        rating: 9.0,
        is_active: true,
        is_featured: true,
        view_count: 145200,
        duration: 9120, // 152 minutes
        release_year: 2008,
        quality: '4K',
        created_at: new Date(Date.now() - 180 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 6 * 86400000).toISOString(),
      },
      {
        id: 103,
        name: 'Inception',
        stream_type: 'vod',
        category_id: 4,
        stream_url: 'https://vod.example.com/inception.mp4',
        thumbnail: 'https://via.placeholder.com/400x600/2c3e50/ecf0f1?text=Inception',
        description: 'A thief who steals corporate secrets through dream-sharing technology.',
        language: 'English',
        country: 'USA',
        genre: 'Sci-Fi',
        rating: 8.8,
        is_active: true,
        is_featured: false,
        view_count: 98450,
        duration: 8880, // 148 minutes
        release_year: 2010,
        quality: 'FHD',
        created_at: new Date(Date.now() - 170 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 7 * 86400000).toISOString(),
      },
      {
        id: 104,
        name: 'Pulp Fiction',
        stream_type: 'vod',
        category_id: 4,
        stream_url: 'https://vod.example.com/pulp-fiction.mp4',
        thumbnail: 'https://via.placeholder.com/400x600/f39c12/2c3e50?text=Pulp+Fiction',
        description: 'The lives of two mob hitmen, a boxer, and a pair of diner bandits intertwine.',
        language: 'English',
        country: 'USA',
        genre: 'Crime',
        rating: 8.9,
        is_active: true,
        is_featured: true,
        view_count: 87230,
        duration: 9240, // 154 minutes
        release_year: 1994,
        quality: 'FHD',
        created_at: new Date(Date.now() - 190 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 8 * 86400000).toISOString(),
      },
      {
        id: 105,
        name: 'Forrest Gump',
        stream_type: 'vod',
        category_id: 4,
        stream_url: 'https://vod.example.com/forrest-gump.mp4',
        thumbnail: 'https://via.placeholder.com/400x600/16a085/ffffff?text=Forrest+Gump',
        description: 'The presidencies of Kennedy and Johnson unfold through the perspective of an Alabama man.',
        language: 'English',
        country: 'USA',
        genre: 'Drama',
        rating: 8.8,
        is_active: true,
        is_featured: false,
        view_count: 112890,
        duration: 8520, // 142 minutes
        release_year: 1994,
        quality: 'FHD',
        created_at: new Date(Date.now() - 195 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 9 * 86400000).toISOString(),
      },
    ];

    const mockStats: StreamStats = {
      total_streams: mockStreams.length,
      live_streams: mockStreams.filter(s => s.stream_type === 'live').length,
      vod_streams: mockStreams.filter(s => s.stream_type === 'vod').length,
      featured_streams: mockStreams.filter(s => s.is_featured).length,
      active_streams: mockStreams.filter(s => s.is_active).length,
      total_views: mockStreams.reduce((sum, s) => sum + s.view_count, 0),
    };

    setStreams(mockStreams);
    setStats(mockStats);
  };

  const handleCreateStream = async () => {
    if (!formData.name || !formData.stream_url) {
      alert('Please fill in all required fields');
      return;
    }

    // In production: await streamAPI.createStream(formData);
    console.log('Creating stream:', formData);

    // Mock: Add to list
    const newStream: Stream = {
      id: streams.length + 1,
      ...formData,
      view_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setStreams([newStream, ...streams]);

    setShowCreateModal(false);
    resetForm();
    alert('Stream created successfully!');
  };

  const handleToggleFeatured = async (stream: Stream) => {
    // In production: await streamAPI.updateStream(stream.id, { is_featured: !stream.is_featured });
    console.log('Toggling featured for stream:', stream.id);

    // Mock: Update stream
    setStreams(streams.map(s =>
      s.id === stream.id ? { ...s, is_featured: !s.is_featured } : s
    ));
  };

  const handleDeleteStream = async (stream: Stream) => {
    if (!confirm(`Delete stream "${stream.name}"?`)) return;

    // In production: await streamAPI.deleteStream(stream.id);
    console.log('Deleting stream:', stream.id);

    // Mock: Remove stream
    setStreams(streams.filter(s => s.id !== stream.id));
    alert('Stream deleted successfully!');
  };

  const resetForm = () => {
    setFormData({
      name: '',
      stream_type: 'live',
      category_id: 1,
      stream_url: '',
      thumbnail: '',
      description: '',
      language: 'English',
      country: 'USA',
      genre: '',
      rating: 0,
      is_active: true,
      is_featured: false,
      duration: 0,
      release_year: new Date().getFullYear(),
      quality: 'FHD',
    });
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const formatViews = (views: number): string => {
    if (views >= 1000000) return `${(views / 1000000).toFixed(1)}M`;
    if (views >= 1000) return `${(views / 1000).toFixed(1)}K`;
    return views.toString();
  };

  const filteredStreams = streams.filter(stream => {
    const matchesType = stream.stream_type === viewMode;
    const matchesSearch = stream.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      stream.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      stream.genre?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesCategory = filterCategory === 'all' || stream.category_id === filterCategory;
    const matchesQuality = filterQuality === 'all' || stream.quality === filterQuality;

    return matchesType && matchesSearch && matchesCategory && matchesQuality;
  });

  return (
    <div className="p-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <PlayCircleIcon className="w-8 h-8 text-blue-500" />
          Stream Management
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          Manage live streams and video on demand content
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Total Streams</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-2">{stats.total_streams}</p>
            </div>
            <PlayCircleIcon className="w-12 h-12 text-blue-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Live Streams</p>
              <p className="text-3xl font-bold text-red-600 dark:text-red-400 mt-2">{stats.live_streams}</p>
            </div>
            <TvIcon className="w-12 h-12 text-red-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">VOD Content</p>
              <p className="text-3xl font-bold text-purple-600 dark:text-purple-400 mt-2">{stats.vod_streams}</p>
            </div>
            <FilmIcon className="w-12 h-12 text-purple-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Total Views</p>
              <p className="text-3xl font-bold text-green-600 dark:text-green-400 mt-2">{formatViews(stats.total_views)}</p>
            </div>
            <StarIconSolid className="w-12 h-12 text-yellow-500" />
          </div>
        </div>
      </div>

      {/* View Mode Tabs */}
      <div className="mb-6 border-b border-gray-200 dark:border-gray-700">
        <div className="flex gap-4">
          <button
            onClick={() => setViewMode('live')}
            className={`pb-4 px-2 border-b-2 font-medium transition-colors flex items-center gap-2 ${
              viewMode === 'live'
                ? 'border-red-500 text-red-600 dark:text-red-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            <TvIcon className="w-5 h-5" />
            Live Streams ({stats.live_streams})
          </button>
          <button
            onClick={() => setViewMode('vod')}
            className={`pb-4 px-2 border-b-2 font-medium transition-colors flex items-center gap-2 ${
              viewMode === 'vod'
                ? 'border-purple-500 text-purple-600 dark:text-purple-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            <FilmIcon className="w-5 h-5" />
            VOD Content ({stats.vod_streams})
          </button>
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
                placeholder="Search streams..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              />
            </div>

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

          <button
            onClick={() => {
              setFormData({ ...formData, stream_type: viewMode });
              setShowCreateModal(true);
            }}
            className="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg flex items-center gap-2 justify-center transition-colors"
          >
            <PlusIcon className="w-5 h-5" />
            Add {viewMode === 'live' ? 'Live Stream' : 'VOD Content'}
          </button>
        </div>
      </div>

      {/* Streams Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {filteredStreams.map((stream) => (
          <div
            key={stream.id}
            className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden hover:shadow-xl transition-shadow cursor-pointer group"
          >
            {/* Thumbnail */}
            <div className="relative aspect-video bg-gray-200 dark:bg-gray-700">
              {stream.thumbnail ? (
                <img
                  src={stream.thumbnail}
                  alt={stream.name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center">
                  {stream.stream_type === 'live' ? (
                    <TvIcon className="w-16 h-16 text-gray-400" />
                  ) : (
                    <FilmIcon className="w-16 h-16 text-gray-400" />
                  )}
                </div>
              )}

              {/* Featured Badge */}
              {stream.is_featured && (
                <div className="absolute top-2 right-2">
                  <div className="bg-yellow-500 text-white px-2 py-1 rounded-full flex items-center gap-1 text-xs font-medium">
                    <StarIconSolid className="w-4 h-4" />
                    Featured
                  </div>
                </div>
              )}

              {/* Type Badge */}
              <div className="absolute top-2 left-2">
                <div className={`px-2 py-1 rounded text-xs font-medium text-white ${
                  stream.stream_type === 'live' ? 'bg-red-500' : 'bg-purple-500'
                }`}>
                  {stream.stream_type === 'live' ? 'LIVE' : 'VOD'}
                </div>
              </div>

              {/* Quality Badge */}
              {stream.quality && (
                <div className="absolute bottom-2 right-2">
                  <div className="bg-black bg-opacity-75 text-white px-2 py-1 rounded text-xs font-medium">
                    {stream.quality}
                  </div>
                </div>
              )}

              {/* Overlay Actions */}
              <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-50 transition-all flex items-center justify-center opacity-0 group-hover:opacity-100">
                <button
                  onClick={() => {
                    setSelectedStream(stream);
                    setShowDetailsModal(true);
                  }}
                  className="bg-white text-gray-900 px-4 py-2 rounded-lg font-medium hover:bg-gray-100 transition-colors"
                >
                  View Details
                </button>
              </div>
            </div>

            {/* Content */}
            <div className="p-4">
              <h3 className="font-semibold text-gray-900 dark:text-white text-lg mb-2 line-clamp-1">
                {stream.name}
              </h3>

              <div className="space-y-2 text-sm text-gray-600 dark:text-gray-400 mb-3">
                {stream.genre && (
                  <div className="flex items-center gap-2">
                    <span className="font-medium">Genre:</span>
                    <span>{stream.genre}</span>
                  </div>
                )}

                {stream.stream_type === 'vod' && stream.release_year && (
                  <div className="flex items-center gap-2">
                    <span className="font-medium">Year:</span>
                    <span>{stream.release_year}</span>
                  </div>
                )}

                {stream.stream_type === 'vod' && stream.duration && (
                  <div className="flex items-center gap-2">
                    <span className="font-medium">Duration:</span>
                    <span>{formatDuration(stream.duration)}</span>
                  </div>
                )}

                {stream.stream_type === 'vod' && stream.rating && (
                  <div className="flex items-center gap-2">
                    <StarIconSolid className="w-4 h-4 text-yellow-500" />
                    <span className="font-medium">{stream.rating.toFixed(1)}/10</span>
                  </div>
                )}

                <div className="flex items-center gap-2">
                  <span className="font-medium">Views:</span>
                  <span>{formatViews(stream.view_count)}</span>
                </div>
              </div>

              {/* Actions */}
              <div className="flex items-center gap-2 pt-3 border-t border-gray-200 dark:border-gray-700">
                <button
                  onClick={() => handleToggleFeatured(stream)}
                  className={`flex-1 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                    stream.is_featured
                      ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 hover:bg-yellow-200 dark:hover:bg-yellow-800'
                      : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600'
                  }`}
                  title={stream.is_featured ? 'Remove from featured' : 'Add to featured'}
                >
                  <StarIcon className="w-4 h-4 mx-auto" />
                </button>
                <button
                  onClick={() => handleDeleteStream(stream)}
                  className="flex-1 px-3 py-2 rounded-lg text-sm font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200 hover:bg-red-200 dark:hover:bg-red-800 transition-colors"
                  title="Delete stream"
                >
                  <TrashIcon className="w-4 h-4 mx-auto" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Empty State */}
      {filteredStreams.length === 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-12 text-center">
          <PlayCircleIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-500 dark:text-gray-400 text-lg">No streams found</p>
          <p className="text-gray-400 dark:text-gray-500 text-sm mt-2">
            Try adjusting your search or filters
          </p>
        </div>
      )}

      {/* Create Stream Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
                Create {formData.stream_type === 'live' ? 'Live Stream' : 'VOD Content'}
              </h2>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Stream Name *
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="e.g., CNN International"
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Stream URL *
                    </label>
                    <input
                      type="url"
                      value={formData.stream_url}
                      onChange={(e) => setFormData({ ...formData, stream_url: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="https://..."
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Quality
                    </label>
                    <select
                      value={formData.quality}
                      onChange={(e) => setFormData({ ...formData, quality: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    >
                      <option value="HD">HD</option>
                      <option value="FHD">Full HD</option>
                      <option value="4K">4K</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Thumbnail URL
                  </label>
                  <input
                    type="url"
                    value={formData.thumbnail}
                    onChange={(e) => setFormData({ ...formData, thumbnail: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="https://..."
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Description
                  </label>
                  <textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="Stream description..."
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Genre
                    </label>
                    <input
                      type="text"
                      value={formData.genre}
                      onChange={(e) => setFormData({ ...formData, genre: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="e.g., News, Sports"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Language
                    </label>
                    <input
                      type="text"
                      value={formData.language}
                      onChange={(e) => setFormData({ ...formData, language: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="e.g., English"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Country
                    </label>
                    <input
                      type="text"
                      value={formData.country}
                      onChange={(e) => setFormData({ ...formData, country: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="e.g., USA"
                    />
                  </div>
                </div>

                {formData.stream_type === 'vod' && (
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Release Year
                      </label>
                      <input
                        type="number"
                        value={formData.release_year}
                        onChange={(e) => setFormData({ ...formData, release_year: parseInt(e.target.value) })}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                        placeholder="2024"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Duration (minutes)
                      </label>
                      <input
                        type="number"
                        value={formData.duration ? Math.floor(formData.duration / 60) : 0}
                        onChange={(e) => setFormData({ ...formData, duration: parseInt(e.target.value) * 60 })}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                        placeholder="120"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Rating (0-10)
                      </label>
                      <input
                        type="number"
                        step="0.1"
                        min="0"
                        max="10"
                        value={formData.rating}
                        onChange={(e) => setFormData({ ...formData, rating: parseFloat(e.target.value) })}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                        placeholder="8.5"
                      />
                    </div>
                  </div>
                )}

                <div className="flex items-center gap-6">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formData.is_active}
                      onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                    />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Active</span>
                  </label>

                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formData.is_featured}
                      onChange={(e) => setFormData({ ...formData, is_featured: e.target.checked })}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                    />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Featured</span>
                  </label>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={handleCreateStream}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Create Stream
                </button>
                <button
                  onClick={() => {
                    setShowCreateModal(false);
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

      {/* Stream Details Modal */}
      {showDetailsModal && selectedStream && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-3xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              {/* Header */}
              <div className="flex items-start justify-between mb-6">
                <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
                  {selectedStream.name}
                </h2>
                <button
                  onClick={() => {
                    setShowDetailsModal(false);
                    setSelectedStream(null);
                  }}
                  className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                >
                  <XMarkIcon className="w-6 h-6" />
                </button>
              </div>

              {/* Thumbnail */}
              {selectedStream.thumbnail && (
                <div className="mb-6 rounded-lg overflow-hidden">
                  <img
                    src={selectedStream.thumbnail}
                    alt={selectedStream.name}
                    className="w-full object-cover"
                  />
                </div>
              )}

              {/* Details */}
              <div className="space-y-4">
                <div>
                  <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Description</h3>
                  <p className="text-gray-900 dark:text-white">
                    {selectedStream.description || 'No description available'}
                  </p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Type</h3>
                    <p className="text-gray-900 dark:text-white capitalize">{selectedStream.stream_type}</p>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Quality</h3>
                    <p className="text-gray-900 dark:text-white">{selectedStream.quality}</p>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Genre</h3>
                    <p className="text-gray-900 dark:text-white">{selectedStream.genre || 'N/A'}</p>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Language</h3>
                    <p className="text-gray-900 dark:text-white">{selectedStream.language || 'N/A'}</p>
                  </div>
                </div>

                {selectedStream.stream_type === 'vod' && (
                  <>
                    <div className="grid grid-cols-3 gap-4">
                      <div>
                        <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Year</h3>
                        <p className="text-gray-900 dark:text-white">{selectedStream.release_year}</p>
                      </div>
                      <div>
                        <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Duration</h3>
                        <p className="text-gray-900 dark:text-white">
                          {selectedStream.duration ? formatDuration(selectedStream.duration) : 'N/A'}
                        </p>
                      </div>
                      <div>
                        <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Rating</h3>
                        <p className="text-gray-900 dark:text-white flex items-center gap-1">
                          <StarIconSolid className="w-5 h-5 text-yellow-500" />
                          {selectedStream.rating?.toFixed(1)}/10
                        </p>
                      </div>
                    </div>
                  </>
                )}

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Views</h3>
                    <p className="text-gray-900 dark:text-white">{formatViews(selectedStream.view_count)}</p>
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Status</h3>
                    <div className="flex items-center gap-2">
                      {selectedStream.is_active ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                          Active
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300">
                          Inactive
                        </span>
                      )}
                      {selectedStream.is_featured && (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
                          Featured
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Stream URL</h3>
                  <p className="text-gray-900 dark:text-white font-mono text-sm break-all">
                    {selectedStream.stream_url}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Streams;
