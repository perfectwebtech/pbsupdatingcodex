import { useEffect, useState } from 'react';
import {
  FilmIcon,
  PlayIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  StarIcon,
  EyeIcon,
  PencilIcon,
  TrashIcon,
  ChevronDownIcon,
  ChevronUpIcon,
} from '@heroicons/react/24/outline';

interface Series {
  id: number;
  name: string;
  description: string;
  category_id?: number;
  category_name?: string;
  cover_url: string;
  backdrop_url: string;
  trailer_url: string;
  rating: number;
  release_year: number;
  genre: string;
  cast: string[];
  director: string;
  producer: string;
  is_active: boolean;
  is_featured: boolean;
  total_seasons: number;
  total_episodes: number;
  view_count: number;
  created_at: string;
}

interface Episode {
  id: number;
  series_id: number;
  series_name: string;
  season: number;
  episode: number;
  title: string;
  description: string;
  stream_url: string;
  thumbnail_url: string;
  duration: number; // in seconds
  air_date?: string;
  rating: number;
  is_active: boolean;
  view_count: number;
  created_at: string;
}

interface SeriesFormData {
  name: string;
  description: string;
  category_id?: number;
  cover_url: string;
  backdrop_url: string;
  trailer_url: string;
  rating: number;
  release_year: number;
  genre: string;
  cast: string[];
  director: string;
  producer: string;
  is_featured: boolean;
}

interface EpisodeFormData {
  series_id: number;
  season: number;
  episode: number;
  title: string;
  description: string;
  stream_url: string;
  thumbnail_url: string;
  duration: number;
  air_date?: string;
  rating: number;
}

export default function Series() {
  const [series, setSeries] = useState<Series[]>([]);
  const [episodes, setEpisodes] = useState<Episode[]>([]);
  const [selectedSeries, setSelectedSeries] = useState<Series | null>(null);
  const [loading, setLoading] = useState(true);
  const [showSeriesModal, setShowSeriesModal] = useState(false);
  const [showEpisodeModal, setShowEpisodeModal] = useState(false);
  const [editingSeries, setEditingSeries] = useState<Series | null>(null);
  const [editingEpisode, setEditingEpisode] = useState<Episode | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [expandedSeasons, setExpandedSeasons] = useState<Set<number>>(new Set());
  const [castInput, setCastInput] = useState('');

  const [seriesFormData, setSeriesFormData] = useState<SeriesFormData>({
    name: '',
    description: '',
    cover_url: '',
    backdrop_url: '',
    trailer_url: '',
    rating: 0,
    release_year: new Date().getFullYear(),
    genre: '',
    cast: [],
    director: '',
    producer: '',
    is_featured: false,
  });

  const [episodeFormData, setEpisodeFormData] = useState<EpisodeFormData>({
    series_id: 0,
    season: 1,
    episode: 1,
    title: '',
    description: '',
    stream_url: '',
    thumbnail_url: '',
    duration: 0,
    rating: 0,
  });

  // Mock data
  const mockSeries: Series[] = [
    {
      id: 1,
      name: 'Breaking Bad',
      description: 'A high school chemistry teacher turned methamphetamine producer.',
      category_id: 1,
      category_name: 'Drama',
      cover_url: 'https://example.com/breaking-bad-cover.jpg',
      backdrop_url: 'https://example.com/breaking-bad-backdrop.jpg',
      trailer_url: 'https://example.com/breaking-bad-trailer.mp4',
      rating: 9.5,
      release_year: 2008,
      genre: 'Crime, Drama, Thriller',
      cast: ['Bryan Cranston', 'Aaron Paul', 'Anna Gunn'],
      director: 'Vince Gilligan',
      producer: 'Mark Johnson',
      is_active: true,
      is_featured: true,
      total_seasons: 5,
      total_episodes: 62,
      view_count: 125000,
      created_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 2,
      name: 'Game of Thrones',
      description: 'Nine noble families fight for control over the lands of Westeros.',
      category_id: 2,
      category_name: 'Fantasy',
      cover_url: 'https://example.com/got-cover.jpg',
      backdrop_url: 'https://example.com/got-backdrop.jpg',
      trailer_url: 'https://example.com/got-trailer.mp4',
      rating: 9.3,
      release_year: 2011,
      genre: 'Adventure, Drama, Fantasy',
      cast: ['Emilia Clarke', 'Kit Harington', 'Peter Dinklage'],
      director: 'David Benioff',
      producer: 'D.B. Weiss',
      is_active: true,
      is_featured: true,
      total_seasons: 8,
      total_episodes: 73,
      view_count: 200000,
      created_at: '2024-01-02T00:00:00Z',
    },
  ];

  const mockEpisodes: Episode[] = [
    {
      id: 1,
      series_id: 1,
      series_name: 'Breaking Bad',
      season: 1,
      episode: 1,
      title: 'Pilot',
      description: 'Walter White, a struggling chemistry teacher, is diagnosed with cancer.',
      stream_url: 'https://example.com/bb-s01e01.m3u8',
      thumbnail_url: 'https://example.com/bb-s01e01-thumb.jpg',
      duration: 3480, // 58 minutes
      air_date: '2008-01-20',
      rating: 9.0,
      is_active: true,
      view_count: 15000,
      created_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 2,
      series_id: 1,
      series_name: 'Breaking Bad',
      season: 1,
      episode: 2,
      title: "Cat's in the Bag...",
      description: 'Walt and Jesse try to dispose of the bodies.',
      stream_url: 'https://example.com/bb-s01e02.m3u8',
      thumbnail_url: 'https://example.com/bb-s01e02-thumb.jpg',
      duration: 2880, // 48 minutes
      air_date: '2008-01-27',
      rating: 8.8,
      is_active: true,
      view_count: 14500,
      created_at: '2024-01-01T00:00:00Z',
    },
  ];

  useEffect(() => {
    loadSeries();
  }, [searchTerm, categoryFilter]);

  const loadSeries = async () => {
    try {
      setLoading(true);
      // In production: const response = await seriesAPI.listSeries({ search: searchTerm, category_id: categoryFilter });
      setTimeout(() => {
        let filtered = mockSeries;
        if (searchTerm) {
          filtered = filtered.filter(s =>
            s.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            s.description.toLowerCase().includes(searchTerm.toLowerCase())
          );
        }
        if (categoryFilter) {
          filtered = filtered.filter(s => s.category_id === parseInt(categoryFilter));
        }
        setSeries(filtered);
        setLoading(false);
      }, 800);
    } catch (error) {
      console.error('Failed to load series:', error);
      setLoading(false);
    }
  };

  const loadEpisodes = async (seriesId: number) => {
    try {
      // In production: const response = await seriesAPI.listEpisodes(seriesId);
      const filtered = mockEpisodes.filter(e => e.series_id === seriesId);
      setEpisodes(filtered);
    } catch (error) {
      console.error('Failed to load episodes:', error);
    }
  };

  const handleCreateSeries = async () => {
    try {
      if (!seriesFormData.name || !seriesFormData.description) {
        alert('Please fill in all required fields');
        return;
      }

      // In production: await seriesAPI.createSeries(seriesFormData);
      console.log('Creating series:', seriesFormData);

      alert('Series created successfully!');
      setShowSeriesModal(false);
      resetSeriesForm();
      loadSeries();
    } catch (error) {
      console.error('Failed to create series:', error);
      alert('Failed to create series');
    }
  };

  const handleUpdateSeries = async () => {
    if (!editingSeries) return;

    try {
      // In production: await seriesAPI.updateSeries(editingSeries.id, seriesFormData);
      console.log('Updating series:', editingSeries.id, seriesFormData);

      alert('Series updated successfully!');
      setShowSeriesModal(false);
      setEditingSeries(null);
      resetSeriesForm();
      loadSeries();
    } catch (error) {
      console.error('Failed to update series:', error);
      alert('Failed to update series');
    }
  };

  const handleDeleteSeries = async (id: number) => {
    if (!confirm('Are you sure you want to delete this series? This will also delete all episodes.')) return;

    try {
      // In production: await seriesAPI.deleteSeries(id);
      console.log('Deleting series:', id);
      alert('Series deleted successfully!');
      loadSeries();
    } catch (error) {
      console.error('Failed to delete series:', error);
      alert('Failed to delete series');
    }
  };

  const handleCreateEpisode = async () => {
    try {
      if (!episodeFormData.title || !episodeFormData.stream_url) {
        alert('Please fill in all required fields');
        return;
      }

      // In production: await seriesAPI.createEpisode(episodeFormData);
      console.log('Creating episode:', episodeFormData);

      alert('Episode created successfully!');
      setShowEpisodeModal(false);
      resetEpisodeForm();
      if (selectedSeries) {
        loadEpisodes(selectedSeries.id);
      }
    } catch (error) {
      console.error('Failed to create episode:', error);
      alert('Failed to create episode');
    }
  };

  const handleUpdateEpisode = async () => {
    if (!editingEpisode) return;

    try {
      // In production: await seriesAPI.updateEpisode(editingEpisode.id, episodeFormData);
      console.log('Updating episode:', editingEpisode.id, episodeFormData);

      alert('Episode updated successfully!');
      setShowEpisodeModal(false);
      setEditingEpisode(null);
      resetEpisodeForm();
      if (selectedSeries) {
        loadEpisodes(selectedSeries.id);
      }
    } catch (error) {
      console.error('Failed to update episode:', error);
      alert('Failed to update episode');
    }
  };

  const handleDeleteEpisode = async (id: number) => {
    if (!confirm('Are you sure you want to delete this episode?')) return;

    try {
      // In production: await seriesAPI.deleteEpisode(id);
      console.log('Deleting episode:', id);
      alert('Episode deleted successfully!');
      if (selectedSeries) {
        loadEpisodes(selectedSeries.id);
      }
    } catch (error) {
      console.error('Failed to delete episode:', error);
      alert('Failed to delete episode');
    }
  };

  const openCreateSeriesModal = () => {
    resetSeriesForm();
    setEditingSeries(null);
    setShowSeriesModal(true);
  };

  const openEditSeriesModal = (series: Series) => {
    setSeriesFormData({
      name: series.name,
      description: series.description,
      category_id: series.category_id,
      cover_url: series.cover_url,
      backdrop_url: series.backdrop_url,
      trailer_url: series.trailer_url,
      rating: series.rating,
      release_year: series.release_year,
      genre: series.genre,
      cast: series.cast,
      director: series.director,
      producer: series.producer,
      is_featured: series.is_featured,
    });
    setEditingSeries(series);
    setShowSeriesModal(true);
  };

  const openCreateEpisodeModal = (series: Series) => {
    resetEpisodeForm();
    setEpisodeFormData({ ...episodeFormData, series_id: series.id });
    setEditingEpisode(null);
    setShowEpisodeModal(true);
  };

  const openEditEpisodeModal = (episode: Episode) => {
    setEpisodeFormData({
      series_id: episode.series_id,
      season: episode.season,
      episode: episode.episode,
      title: episode.title,
      description: episode.description,
      stream_url: episode.stream_url,
      thumbnail_url: episode.thumbnail_url,
      duration: episode.duration,
      air_date: episode.air_date,
      rating: episode.rating,
    });
    setEditingEpisode(episode);
    setShowEpisodeModal(true);
  };

  const viewSeriesDetails = (series: Series) => {
    setSelectedSeries(series);
    loadEpisodes(series.id);
  };

  const resetSeriesForm = () => {
    setSeriesFormData({
      name: '',
      description: '',
      cover_url: '',
      backdrop_url: '',
      trailer_url: '',
      rating: 0,
      release_year: new Date().getFullYear(),
      genre: '',
      cast: [],
      director: '',
      producer: '',
      is_featured: false,
    });
    setCastInput('');
  };

  const resetEpisodeForm = () => {
    setEpisodeFormData({
      series_id: 0,
      season: 1,
      episode: 1,
      title: '',
      description: '',
      stream_url: '',
      thumbnail_url: '',
      duration: 0,
      rating: 0,
    });
  };

  const addCastMember = () => {
    if (castInput.trim()) {
      setSeriesFormData({
        ...seriesFormData,
        cast: [...seriesFormData.cast, castInput.trim()],
      });
      setCastInput('');
    }
  };

  const removeCastMember = (index: number) => {
    setSeriesFormData({
      ...seriesFormData,
      cast: seriesFormData.cast.filter((_, i) => i !== index),
    });
  };

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
  };

  const toggleSeason = (season: number) => {
    const newExpanded = new Set(expandedSeasons);
    if (newExpanded.has(season)) {
      newExpanded.delete(season);
    } else {
      newExpanded.add(season);
    }
    setExpandedSeasons(newExpanded);
  };

  const groupEpisodesBySeason = (episodes: Episode[]) => {
    const grouped: { [key: number]: Episode[] } = {};
    episodes.forEach(ep => {
      if (!grouped[ep.season]) {
        grouped[ep.season] = [];
      }
      grouped[ep.season].push(ep);
    });
    return grouped;
  };

  if (selectedSeries) {
    const episodesBySeason = groupEpisodesBySeason(episodes);
    const seasons = Object.keys(episodesBySeason).map(Number).sort((a, b) => a - b);

    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="card p-6">
          <button
            onClick={() => setSelectedSeries(null)}
            className="btn-secondary mb-4"
          >
            ← Back to Series List
          </button>

          <div className="flex items-start justify-between">
            <div className="flex-1">
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
                {selectedSeries.name}
              </h1>
              <p className="text-gray-600 dark:text-gray-400 mb-4">{selectedSeries.description}</p>

              <div className="flex items-center space-x-6 text-sm">
                <div className="flex items-center">
                  <StarIcon className="h-5 w-5 text-yellow-400 mr-1" />
                  <span className="font-semibold">{selectedSeries.rating}</span>
                </div>
                <span className="text-gray-500">{selectedSeries.release_year}</span>
                <span className="text-gray-500">{selectedSeries.genre}</span>
                <span className="text-gray-500">
                  {selectedSeries.total_seasons} Seasons • {selectedSeries.total_episodes} Episodes
                </span>
                <span className="text-gray-500">
                  <EyeIcon className="h-4 w-4 inline mr-1" />
                  {selectedSeries.view_count.toLocaleString()} views
                </span>
              </div>
            </div>

            <button
              onClick={() => openCreateEpisodeModal(selectedSeries)}
              className="btn-primary"
            >
              <PlusIcon className="h-5 w-5 mr-2" />
              Add Episode
            </button>
          </div>
        </div>

        {/* Episodes by Season */}
        <div className="space-y-4">
          {seasons.map(season => (
            <div key={season} className="card overflow-hidden">
              <button
                onClick={() => toggleSeason(season)}
                className="w-full px-6 py-4 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-dark-800 transition-colors"
              >
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                  Season {season} ({episodesBySeason[season].length} episodes)
                </h3>
                {expandedSeasons.has(season) ? (
                  <ChevronUpIcon className="h-5 w-5 text-gray-500" />
                ) : (
                  <ChevronDownIcon className="h-5 w-5 text-gray-500" />
                )}
              </button>

              {expandedSeasons.has(season) && (
                <div className="border-t border-gray-200 dark:border-dark-700">
                  {episodesBySeason[season].map(episode => (
                    <div
                      key={episode.id}
                      className="px-6 py-4 border-b border-gray-200 dark:border-dark-700 last:border-b-0 hover:bg-gray-50 dark:hover:bg-dark-800 transition-colors"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3 mb-2">
                            <span className="text-sm font-semibold text-primary-600 dark:text-primary-400">
                              Episode {episode.episode}
                            </span>
                            <h4 className="text-lg font-semibold text-gray-900 dark:text-white">
                              {episode.title}
                            </h4>
                            {episode.rating > 0 && (
                              <div className="flex items-center">
                                <StarIcon className="h-4 w-4 text-yellow-400 mr-1" />
                                <span className="text-sm font-medium">{episode.rating}</span>
                              </div>
                            )}
                          </div>
                          <p className="text-gray-600 dark:text-gray-400 text-sm mb-2">
                            {episode.description}
                          </p>
                          <div className="flex items-center space-x-4 text-xs text-gray-500">
                            <span>{formatDuration(episode.duration)}</span>
                            {episode.air_date && <span>Aired: {new Date(episode.air_date).toLocaleDateString()}</span>}
                            <span>{episode.view_count.toLocaleString()} views</span>
                          </div>
                        </div>

                        <div className="flex items-center space-x-2 ml-4">
                          <button
                            onClick={() => openEditEpisodeModal(episode)}
                            className="btn-icon"
                            title="Edit Episode"
                          >
                            <PencilIcon className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => handleDeleteEpisode(episode.id)}
                            className="btn-icon text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30"
                            title="Delete Episode"
                          >
                            <TrashIcon className="h-4 w-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Episode Modal */}
        {showEpisodeModal && (
          <div className="modal-overlay">
            <div className="modal-content max-w-2xl">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
                {editingEpisode ? 'Edit Episode' : 'Add New Episode'}
              </h2>

              <form className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="label">Season *</label>
                    <input
                      type="number"
                      min="1"
                      className="input"
                      value={episodeFormData.season || ''}
                      onChange={(e) => setEpisodeFormData({ ...episodeFormData, season: parseInt(e.target.value) })}
                      required
                    />
                  </div>

                  <div>
                    <label className="label">Episode Number *</label>
                    <input
                      type="number"
                      min="1"
                      className="input"
                      value={episodeFormData.episode || ''}
                      onChange={(e) => setEpisodeFormData({ ...episodeFormData, episode: parseInt(e.target.value) })}
                      required
                    />
                  </div>
                </div>

                <div>
                  <label className="label">Episode Title *</label>
                  <input
                    type="text"
                    className="input"
                    value={episodeFormData.title}
                    onChange={(e) => setEpisodeFormData({ ...episodeFormData, title: e.target.value })}
                    required
                  />
                </div>

                <div>
                  <label className="label">Description</label>
                  <textarea
                    className="input"
                    rows={3}
                    value={episodeFormData.description}
                    onChange={(e) => setEpisodeFormData({ ...episodeFormData, description: e.target.value })}
                  />
                </div>

                <div>
                  <label className="label">Stream URL *</label>
                  <input
                    type="url"
                    className="input"
                    placeholder="https://cdn.example.com/series/episode.m3u8"
                    value={episodeFormData.stream_url}
                    onChange={(e) => setEpisodeFormData({ ...episodeFormData, stream_url: e.target.value })}
                    required
                  />
                </div>

                <div>
                  <label className="label">Thumbnail URL</label>
                  <input
                    type="url"
                    className="input"
                    placeholder="https://cdn.example.com/thumbnails/episode.jpg"
                    value={episodeFormData.thumbnail_url}
                    onChange={(e) => setEpisodeFormData({ ...episodeFormData, thumbnail_url: e.target.value })}
                  />
                </div>

                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <label className="label">Duration (seconds)</label>
                    <input
                      type="number"
                      min="0"
                      className="input"
                      value={episodeFormData.duration || ''}
                      onChange={(e) => setEpisodeFormData({ ...episodeFormData, duration: parseInt(e.target.value) })}
                    />
                  </div>

                  <div>
                    <label className="label">Air Date</label>
                    <input
                      type="date"
                      className="input"
                      value={episodeFormData.air_date || ''}
                      onChange={(e) => setEpisodeFormData({ ...episodeFormData, air_date: e.target.value })}
                    />
                  </div>

                  <div>
                    <label className="label">Rating</label>
                    <input
                      type="number"
                      min="0"
                      max="10"
                      step="0.1"
                      className="input"
                      value={episodeFormData.rating || ''}
                      onChange={(e) => setEpisodeFormData({ ...episodeFormData, rating: parseFloat(e.target.value) })}
                    />
                  </div>
                </div>

                <div className="flex justify-end space-x-3 pt-4">
                  <button
                    type="button"
                    onClick={() => {
                      setShowEpisodeModal(false);
                      setEditingEpisode(null);
                    }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={editingEpisode ? handleUpdateEpisode : handleCreateEpisode}
                    className="btn-primary"
                  >
                    {editingEpisode ? 'Update Episode' : 'Create Episode'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Filters and Actions */}
      <div className="card p-6">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div className="flex-1 flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search series..."
                className="input pl-10 w-full"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            <div className="relative sm:w-48">
              <FunnelIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
              <select
                className="input pl-10 w-full"
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value)}
              >
                <option value="">All Categories</option>
                <option value="1">Drama</option>
                <option value="2">Fantasy</option>
                <option value="3">Action</option>
                <option value="4">Comedy</option>
              </select>
            </div>
          </div>

          <button
            onClick={openCreateSeriesModal}
            className="btn-primary"
          >
            <PlusIcon className="h-5 w-5 mr-2" />
            Add Series
          </button>
        </div>
      </div>

      {/* Series Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {loading ? (
          Array(8).fill(0).map((_, i) => (
            <div key={i} className="card overflow-hidden">
              <div className="h-64 skeleton" />
              <div className="p-4 space-y-3">
                <div className="h-6 skeleton rounded" />
                <div className="h-4 skeleton rounded w-3/4" />
              </div>
            </div>
          ))
        ) : series.length === 0 ? (
          <div className="col-span-full text-center py-12 text-gray-500">
            <FilmIcon className="h-16 w-16 mx-auto mb-3 opacity-30" />
            <p>No series found</p>
          </div>
        ) : (
          series.map((item) => (
            <div key={item.id} className="card overflow-hidden group cursor-pointer">
              <div
                className="relative h-64 bg-gradient-to-br from-gray-800 to-gray-900 overflow-hidden"
                onClick={() => viewSeriesDetails(item)}
              >
                {item.cover_url ? (
                  <img
                    src={item.cover_url}
                    alt={item.name}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                    onError={(e) => {
                      e.currentTarget.style.display = 'none';
                    }}
                  />
                ) : (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <FilmIcon className="h-16 w-16 text-gray-600" />
                  </div>
                )}

                <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                  <div className="absolute bottom-4 left-4 right-4 flex items-center justify-center space-x-2">
                    <PlayIcon className="h-10 w-10 text-white" />
                  </div>
                </div>

                {item.is_featured && (
                  <div className="absolute top-4 right-4 bg-primary-600 text-white px-2 py-1 rounded-lg text-xs font-semibold">
                    Featured
                  </div>
                )}

                <div className="absolute top-4 left-4 flex items-center space-x-2">
                  <div className="flex items-center bg-black/60 backdrop-blur-sm px-2 py-1 rounded-lg">
                    <StarIcon className="h-4 w-4 text-yellow-400 mr-1" />
                    <span className="text-white text-sm font-semibold">{item.rating}</span>
                  </div>
                </div>
              </div>

              <div className="p-4">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1 truncate">
                  {item.name}
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-3 line-clamp-2">
                  {item.description}
                </p>

                <div className="flex items-center justify-between text-xs text-gray-500 mb-3">
                  <span>{item.release_year}</span>
                  <span>{item.total_seasons} Seasons</span>
                  <span>{item.total_episodes} Episodes</span>
                </div>

                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => viewSeriesDetails(item)}
                    className="flex-1 btn-primary py-2 text-sm"
                  >
                    <EyeIcon className="h-4 w-4 mr-1" />
                    View
                  </button>
                  <button
                    onClick={() => openEditSeriesModal(item)}
                    className="btn-icon"
                    title="Edit"
                  >
                    <PencilIcon className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => handleDeleteSeries(item.id)}
                    className="btn-icon text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30"
                    title="Delete"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Series Create/Edit Modal */}
      {showSeriesModal && (
        <div className="modal-overlay">
          <div className="modal-content max-w-4xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
              {editingSeries ? 'Edit Series' : 'Add New Series'}
            </h2>

            <form className="space-y-6">
              <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2">
                  <label className="label">Series Name *</label>
                  <input
                    type="text"
                    className="input"
                    value={seriesFormData.name}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, name: e.target.value })}
                    required
                  />
                </div>

                <div className="col-span-2">
                  <label className="label">Description *</label>
                  <textarea
                    className="input"
                    rows={3}
                    value={seriesFormData.description}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, description: e.target.value })}
                    required
                  />
                </div>

                <div>
                  <label className="label">Genre</label>
                  <input
                    type="text"
                    className="input"
                    placeholder="Drama, Thriller, Crime"
                    value={seriesFormData.genre}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, genre: e.target.value })}
                  />
                </div>

                <div>
                  <label className="label">Release Year</label>
                  <input
                    type="number"
                    className="input"
                    min="1900"
                    max={new Date().getFullYear() + 5}
                    value={seriesFormData.release_year || ''}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, release_year: parseInt(e.target.value) })}
                  />
                </div>

                <div>
                  <label className="label">Director</label>
                  <input
                    type="text"
                    className="input"
                    value={seriesFormData.director}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, director: e.target.value })}
                  />
                </div>

                <div>
                  <label className="label">Producer</label>
                  <input
                    type="text"
                    className="input"
                    value={seriesFormData.producer}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, producer: e.target.value })}
                  />
                </div>

                <div className="col-span-2">
                  <label className="label">Cast Members</label>
                  <div className="flex space-x-2 mb-2">
                    <input
                      type="text"
                      className="input flex-1"
                      placeholder="Add cast member name"
                      value={castInput}
                      onChange={(e) => setCastInput(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addCastMember())}
                    />
                    <button type="button" onClick={addCastMember} className="btn-primary">
                      Add
                    </button>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {seriesFormData.cast.map((member, index) => (
                      <span
                        key={index}
                        className="inline-flex items-center px-3 py-1 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 rounded-full text-sm"
                      >
                        {member}
                        <button
                          type="button"
                          onClick={() => removeCastMember(index)}
                          className="ml-2 text-primary-600 hover:text-primary-800"
                        >
                          ×
                        </button>
                      </span>
                    ))}
                  </div>
                </div>

                <div className="col-span-2">
                  <label className="label">Cover Image URL</label>
                  <input
                    type="url"
                    className="input"
                    placeholder="https://cdn.example.com/cover.jpg"
                    value={seriesFormData.cover_url}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, cover_url: e.target.value })}
                  />
                </div>

                <div className="col-span-2">
                  <label className="label">Backdrop Image URL</label>
                  <input
                    type="url"
                    className="input"
                    placeholder="https://cdn.example.com/backdrop.jpg"
                    value={seriesFormData.backdrop_url}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, backdrop_url: e.target.value })}
                  />
                </div>

                <div className="col-span-2">
                  <label className="label">Trailer URL</label>
                  <input
                    type="url"
                    className="input"
                    placeholder="https://cdn.example.com/trailer.mp4"
                    value={seriesFormData.trailer_url}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, trailer_url: e.target.value })}
                  />
                </div>

                <div>
                  <label className="label">Rating (0-10)</label>
                  <input
                    type="number"
                    min="0"
                    max="10"
                    step="0.1"
                    className="input"
                    value={seriesFormData.rating || ''}
                    onChange={(e) => setSeriesFormData({ ...seriesFormData, rating: parseFloat(e.target.value) })}
                  />
                </div>

                <div className="flex items-center">
                  <label className="flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      className="mr-2"
                      checked={seriesFormData.is_featured}
                      onChange={(e) => setSeriesFormData({ ...seriesFormData, is_featured: e.target.checked })}
                    />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      Feature this series
                    </span>
                  </label>
                </div>
              </div>

              <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200 dark:border-dark-700">
                <button
                  type="button"
                  onClick={() => {
                    setShowSeriesModal(false);
                    setEditingSeries(null);
                  }}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={editingSeries ? handleUpdateSeries : handleCreateSeries}
                  className="btn-primary"
                >
                  {editingSeries ? 'Update Series' : 'Create Series'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
