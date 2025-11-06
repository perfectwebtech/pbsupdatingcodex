import { useEffect, useState } from 'react';
import {
  CalendarIcon,
  ClockIcon,
  TvIcon,
  ArrowPathIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  PlayIcon,
  DocumentArrowUpIcon,
  CheckCircleIcon,
  XCircleIcon,
  EyeIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';

interface EPGProgram {
  id: number;
  stream_id: number;
  stream_name: string;
  title: string;
  description: string;
  category: string;
  start_time: string;
  end_time: string;
  duration: number;
  image_url?: string;
  icon_url?: string;
  episode_number?: number;
  season_number?: number;
  year?: number;
  rating?: string;
  directors?: string[];
  actors?: string[];
  country?: string;
  language?: string;
  is_live: boolean;
  is_repeat: boolean;
  is_premiere: boolean;
  status?: 'now_playing' | 'upcoming' | 'past';
  created_at: string;
}

interface EPGSource {
  id: number;
  name: string;
  source_type: 'xmltv' | 'json' | 'api' | 'manual';
  url?: string;
  update_interval: number;
  is_active: boolean;
  last_sync?: string;
  last_sync_status?: string;
  last_error?: string;
  programs_imported: number;
  programs_updated: number;
  programs_failed: number;
  created_at: string;
}

interface EPGImportLog {
  id: number;
  source_id?: number;
  import_type: string;
  status: 'success' | 'failed' | 'in_progress';
  programs_processed: number;
  programs_created: number;
  programs_updated: number;
  programs_deleted: number;
  programs_failed: number;
  started_at: string;
  completed_at?: string;
  duration_seconds?: number;
  error_message?: string;
  created_at: string;
}

interface ProgramFormData {
  stream_id: number;
  title: string;
  description: string;
  category: string;
  start_time: string;
  end_time: string;
  image_url: string;
  episode_number?: number;
  season_number?: number;
  year?: number;
  rating: string;
  directors: string[];
  actors: string[];
  country: string;
  language: string;
  is_live: boolean;
  is_repeat: boolean;
  is_premiere: boolean;
}

interface SourceFormData {
  name: string;
  source_type: 'xmltv' | 'json' | 'api' | 'manual';
  url: string;
  update_interval: number;
}

type ViewMode = 'programs' | 'sources' | 'imports';

export default function EPG() {
  const [viewMode, setViewMode] = useState<ViewMode>('programs');
  const [programs, setPrograms] = useState<EPGProgram[]>([]);
  const [sources, setSources] = useState<EPGSource[]>([]);
  const [importLogs, setImportLogs] = useState<EPGImportLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [showProgramModal, setShowProgramModal] = useState(false);
  const [showSourceModal, setShowSourceModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [selectedProgram, setSelectedProgram] = useState<EPGProgram | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [streamFilter, setStreamFilter] = useState('');
  const [dateFilter, setDateFilter] = useState(new Date().toISOString().split('T')[0]);

  const [programFormData, setProgramFormData] = useState<ProgramFormData>({
    stream_id: 0,
    title: '',
    description: '',
    category: '',
    start_time: '',
    end_time: '',
    image_url: '',
    rating: '',
    directors: [],
    actors: [],
    country: '',
    language: '',
    is_live: false,
    is_repeat: false,
    is_premiere: false,
  });

  const [sourceFormData, setSourceFormData] = useState<SourceFormData>({
    name: '',
    source_type: 'xmltv',
    url: '',
    update_interval: 3600,
  });

  // Mock data
  const mockPrograms: EPGProgram[] = [
    {
      id: 1,
      stream_id: 1,
      stream_name: 'HBO',
      title: 'Game of Thrones',
      description: 'The final battle begins',
      category: 'Drama',
      start_time: new Date().toISOString(),
      end_time: new Date(Date.now() + 3600000).toISOString(),
      duration: 3600,
      image_url: 'https://example.com/got.jpg',
      season_number: 8,
      episode_number: 6,
      year: 2019,
      rating: 'TV-MA',
      directors: ['David Benioff', 'D.B. Weiss'],
      actors: ['Emilia Clarke', 'Kit Harington'],
      country: 'USA',
      language: 'English',
      is_live: false,
      is_repeat: false,
      is_premiere: true,
      status: 'now_playing',
      created_at: new Date().toISOString(),
    },
    {
      id: 2,
      stream_id: 2,
      stream_name: 'BBC One',
      title: 'News at 10',
      description: 'Latest news and current affairs',
      category: 'News',
      start_time: new Date(Date.now() + 7200000).toISOString(),
      end_time: new Date(Date.now() + 9000000).toISOString(),
      duration: 1800,
      is_live: true,
      is_repeat: false,
      is_premiere: false,
      country: 'UK',
      language: 'English',
      status: 'upcoming',
      created_at: new Date().toISOString(),
    },
  ];

  const mockSources: EPGSource[] = [
    {
      id: 1,
      name: 'Main XMLTV Source',
      source_type: 'xmltv',
      url: 'https://example.com/epg.xml',
      update_interval: 3600,
      is_active: true,
      last_sync: new Date(Date.now() - 1800000).toISOString(),
      last_sync_status: 'success',
      programs_imported: 1250,
      programs_updated: 350,
      programs_failed: 5,
      created_at: new Date().toISOString(),
    },
    {
      id: 2,
      name: 'Backup EPG Source',
      source_type: 'json',
      url: 'https://api.example.com/epg',
      update_interval: 7200,
      is_active: true,
      last_sync: new Date(Date.now() - 3600000).toISOString(),
      last_sync_status: 'failed',
      last_error: 'Connection timeout',
      programs_imported: 800,
      programs_updated: 150,
      programs_failed: 25,
      created_at: new Date().toISOString(),
    },
  ];

  const mockImportLogs: EPGImportLog[] = [
    {
      id: 1,
      source_id: 1,
      import_type: 'full',
      status: 'success',
      programs_processed: 1500,
      programs_created: 850,
      programs_updated: 620,
      programs_deleted: 30,
      programs_failed: 0,
      started_at: new Date(Date.now() - 1800000).toISOString(),
      completed_at: new Date(Date.now() - 1620000).toISOString(),
      duration_seconds: 180,
      created_at: new Date(Date.now() - 1800000).toISOString(),
    },
    {
      id: 2,
      source_id: 2,
      import_type: 'incremental',
      status: 'failed',
      programs_processed: 500,
      programs_created: 0,
      programs_updated: 0,
      programs_deleted: 0,
      programs_failed: 500,
      started_at: new Date(Date.now() - 3600000).toISOString(),
      error_message: 'Failed to connect to EPG source',
      created_at: new Date(Date.now() - 3600000).toISOString(),
    },
  ];

  useEffect(() => {
    loadData();
  }, [viewMode, dateFilter, streamFilter]);

  const loadData = async () => {
    try {
      setLoading(true);
      // In production: await epgAPI.listPrograms({ date: dateFilter, stream_id: streamFilter });
      setTimeout(() => {
        if (viewMode === 'programs') {
          setPrograms(mockPrograms);
        } else if (viewMode === 'sources') {
          setSources(mockSources);
        } else {
          setImportLogs(mockImportLogs);
        }
        setLoading(false);
      }, 800);
    } catch (error) {
      console.error('Failed to load data:', error);
      setLoading(false);
    }
  };

  const handleCreateProgram = async () => {
    try {
      if (!programFormData.title || !programFormData.stream_id) {
        alert('Please fill in all required fields');
        return;
      }

      // In production: await epgAPI.createProgram(programFormData);
      console.log('Creating program:', programFormData);

      alert('EPG program created successfully!');
      setShowProgramModal(false);
      resetProgramForm();
      loadData();
    } catch (error) {
      console.error('Failed to create program:', error);
      alert('Failed to create program');
    }
  };

  const handleCreateSource = async () => {
    try {
      if (!sourceFormData.name || !sourceFormData.url) {
        alert('Please fill in all required fields');
        return;
      }

      // In production: await epgAPI.createSource(sourceFormData);
      console.log('Creating source:', sourceFormData);

      alert('EPG source created successfully!');
      setShowSourceModal(false);
      resetSourceForm();
      loadData();
    } catch (error) {
      console.error('Failed to create source:', error);
      alert('Failed to create source');
    }
  };

  const handleSyncSource = async (sourceId: number) => {
    if (!confirm('Start EPG import from this source?')) return;

    try {
      // In production: await epgAPI.syncSource(sourceId);
      console.log('Syncing source:', sourceId);
      alert('EPG sync started! Check import history for progress.');
      loadData();
    } catch (error) {
      console.error('Failed to sync source:', error);
      alert('Failed to start sync');
    }
  };

  const handleDeleteProgram = async (id: number) => {
    if (!confirm('Are you sure you want to delete this program?')) return;

    try {
      // In production: await epgAPI.deleteProgram(id);
      console.log('Deleting program:', id);
      alert('Program deleted successfully!');
      loadData();
    } catch (error) {
      console.error('Failed to delete program:', error);
      alert('Failed to delete program');
    }
  };

  const handleCleanup = async () => {
    const days = prompt('Remove programs older than how many days?', '30');
    if (!days) return;

    try {
      // In production: await epgAPI.cleanup(parseInt(days));
      console.log('Cleaning up programs older than', days, 'days');
      alert(`Old programs cleaned up successfully!`);
      loadData();
    } catch (error) {
      console.error('Failed to cleanup:', error);
      alert('Failed to cleanup');
    }
  };

  const resetProgramForm = () => {
    setProgramFormData({
      stream_id: 0,
      title: '',
      description: '',
      category: '',
      start_time: '',
      end_time: '',
      image_url: '',
      rating: '',
      directors: [],
      actors: [],
      country: '',
      language: '',
      is_live: false,
      is_repeat: false,
      is_premiere: false,
    });
  };

  const resetSourceForm = () => {
    setSourceFormData({
      name: '',
      source_type: 'xmltv',
      url: '',
      update_interval: 3600,
    });
  };

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
  };

  const formatDateTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusBadge = (status?: string) => {
    const badges = {
      now_playing: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-700 dark:text-green-400', label: 'Live Now' },
      upcoming: { bg: 'bg-blue-100 dark:bg-blue-900/30', text: 'text-blue-700 dark:text-blue-400', label: 'Upcoming' },
      past: { bg: 'bg-gray-100 dark:bg-gray-900/30', text: 'text-gray-700 dark:text-gray-400', label: 'Ended' },
    };

    if (!status) return null;
    const badge = badges[status as keyof typeof badges];
    if (!badge) return null;

    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${badge.bg} ${badge.text}`}>
        {badge.label}
      </span>
    );
  };

  const getSyncStatusBadge = (status?: string) => {
    if (status === 'success') {
      return <CheckCircleIcon className="h-5 w-5 text-green-500" />;
    } else if (status === 'failed') {
      return <XCircleIcon className="h-5 w-5 text-red-500" />;
    }
    return <ClockIcon className="h-5 w-5 text-gray-400" />;
  };

  return (
    <div className="space-y-6">
      {/* View Mode Tabs */}
      <div className="card p-6">
        <div className="flex items-center justify-between">
          <div className="flex space-x-2">
            <button
              onClick={() => setViewMode('programs')}
              className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                viewMode === 'programs'
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 dark:bg-dark-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-dark-700'
              }`}
            >
              <CalendarIcon className="h-5 w-5 inline mr-2" />
              Programs
            </button>
            <button
              onClick={() => setViewMode('sources')}
              className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                viewMode === 'sources'
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 dark:bg-dark-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-dark-700'
              }`}
            >
              <TvIcon className="h-5 w-5 inline mr-2" />
              Sources
            </button>
            <button
              onClick={() => setViewMode('imports')}
              className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                viewMode === 'imports'
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 dark:bg-dark-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-dark-700'
              }`}
            >
              <DocumentArrowUpIcon className="h-5 w-5 inline mr-2" />
              Import History
            </button>
          </div>

          <div className="flex space-x-2">
            {viewMode === 'programs' && (
              <>
                <button onClick={handleCleanup} className="btn-secondary">
                  <TrashIcon className="h-5 w-5 mr-2" />
                  Cleanup Old
                </button>
                <button onClick={() => setShowProgramModal(true)} className="btn-primary">
                  <PlusIcon className="h-5 w-5 mr-2" />
                  Add Program
                </button>
              </>
            )}
            {viewMode === 'sources' && (
              <button onClick={() => setShowSourceModal(true)} className="btn-primary">
                <PlusIcon className="h-5 w-5 mr-2" />
                Add Source
              </button>
            )}
          </div>
        </div>

        {/* Filters */}
        {viewMode === 'programs' && (
          <div className="mt-4 flex space-x-3">
            <div className="relative flex-1">
              <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search programs..."
                className="input pl-10 w-full"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            <input
              type="date"
              className="input w-48"
              value={dateFilter}
              onChange={(e) => setDateFilter(e.target.value)}
            />

            <select
              className="input w-48"
              value={streamFilter}
              onChange={(e) => setStreamFilter(e.target.value)}
            >
              <option value="">All Streams</option>
              <option value="1">HBO</option>
              <option value="2">BBC One</option>
              <option value="3">ESPN</option>
            </select>
          </div>
        )}
      </div>

      {/* Programs View */}
      {viewMode === 'programs' && (
        <div className="card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Stream</th>
                  <th>Program</th>
                  <th>Category</th>
                  <th>Duration</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  Array(5).fill(0).map((_, i) => (
                    <tr key={i}>
                      <td colSpan={7}>
                        <div className="h-12 skeleton rounded" />
                      </td>
                    </tr>
                  ))
                ) : programs.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="text-center py-12 text-gray-500">
                      <CalendarIcon className="h-12 w-12 mx-auto mb-3 opacity-30" />
                      <p>No programs found</p>
                    </td>
                  </tr>
                ) : (
                  programs.map((program) => (
                    <tr key={program.id}>
                      <td>
                        <div className="text-sm">
                          <div className="font-medium">{formatDateTime(program.start_time)}</div>
                          <div className="text-gray-500">{formatDateTime(program.end_time)}</div>
                        </div>
                      </td>
                      <td>
                        <div className="font-medium text-gray-900 dark:text-white">
                          {program.stream_name}
                        </div>
                      </td>
                      <td>
                        <div>
                          <div className="font-semibold text-gray-900 dark:text-white">
                            {program.title}
                          </div>
                          {program.description && (
                            <div className="text-sm text-gray-500 line-clamp-1">
                              {program.description}
                            </div>
                          )}
                          {(program.season_number || program.episode_number) && (
                            <div className="text-xs text-primary-600 dark:text-primary-400 mt-1">
                              S{program.season_number} E{program.episode_number}
                            </div>
                          )}
                        </div>
                      </td>
                      <td>
                        <span className="text-sm text-gray-600 dark:text-gray-400">
                          {program.category}
                        </span>
                      </td>
                      <td className="text-sm">{formatDuration(program.duration)}</td>
                      <td>{getStatusBadge(program.status)}</td>
                      <td>
                        <div className="flex items-center space-x-2">
                          <button
                            onClick={() => {
                              setSelectedProgram(program);
                              setShowViewModal(true);
                            }}
                            className="btn-icon"
                            title="View Details"
                          >
                            <EyeIcon className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => handleDeleteProgram(program.id)}
                            className="btn-icon text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30"
                            title="Delete"
                          >
                            <TrashIcon className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Sources View */}
      {viewMode === 'sources' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {loading ? (
            Array(4).fill(0).map((_, i) => (
              <div key={i} className="card p-6">
                <div className="h-32 skeleton rounded" />
              </div>
            ))
          ) : (
            sources.map((source) => (
              <div key={source.id} className="card p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">
                      {source.name}
                    </h3>
                    <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                      <span className="px-2 py-0.5 bg-gray-100 dark:bg-dark-800 rounded">
                        {source.source_type.toUpperCase()}
                      </span>
                      <span>•</span>
                      <span>Updates every {source.update_interval / 60} min</span>
                    </div>
                  </div>
                  {getSyncStatusBadge(source.last_sync_status)}
                </div>

                {source.url && (
                  <div className="mb-4">
                    <p className="text-sm text-gray-500 dark:text-gray-400 truncate">
                      {source.url}
                    </p>
                  </div>
                )}

                <div className="grid grid-cols-3 gap-4 mb-4">
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Imported</p>
                    <p className="text-lg font-semibold text-green-600 dark:text-green-400">
                      {source.programs_imported}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Updated</p>
                    <p className="text-lg font-semibold text-blue-600 dark:text-blue-400">
                      {source.programs_updated}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Failed</p>
                    <p className="text-lg font-semibold text-red-600 dark:text-red-400">
                      {source.programs_failed}
                    </p>
                  </div>
                </div>

                {source.last_sync && (
                  <p className="text-xs text-gray-500 dark:text-gray-400 mb-4">
                    Last sync: {new Date(source.last_sync).toLocaleString()}
                  </p>
                )}

                {source.last_error && (
                  <div className="mb-4 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-xs text-red-700 dark:text-red-400">
                    {source.last_error}
                  </div>
                )}

                <button
                  onClick={() => handleSyncSource(source.id)}
                  className="w-full btn-primary"
                  disabled={!source.is_active}
                >
                  <ArrowPathIcon className="h-5 w-5 mr-2" />
                  Sync Now
                </button>
              </div>
            ))
          )}
        </div>
      )}

      {/* Import History View */}
      {viewMode === 'imports' && (
        <div className="card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th>Started</th>
                  <th>Type</th>
                  <th>Status</th>
                  <th>Processed</th>
                  <th>Created</th>
                  <th>Updated</th>
                  <th>Failed</th>
                  <th>Duration</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  Array(5).fill(0).map((_, i) => (
                    <tr key={i}>
                      <td colSpan={8}>
                        <div className="h-12 skeleton rounded" />
                      </td>
                    </tr>
                  ))
                ) : (
                  importLogs.map((log) => (
                    <tr key={log.id}>
                      <td className="text-sm">
                        {new Date(log.started_at).toLocaleString()}
                      </td>
                      <td>
                        <span className="text-sm px-2 py-0.5 bg-gray-100 dark:bg-dark-800 rounded">
                          {log.import_type}
                        </span>
                      </td>
                      <td>
                        {log.status === 'success' && (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">
                            <CheckCircleIcon className="h-3 w-3 mr-1" />
                            Success
                          </span>
                        )}
                        {log.status === 'failed' && (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">
                            <XCircleIcon className="h-3 w-3 mr-1" />
                            Failed
                          </span>
                        )}
                        {log.status === 'in_progress' && (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400">
                            <ArrowPathIcon className="h-3 w-3 mr-1 animate-spin" />
                            In Progress
                          </span>
                        )}
                      </td>
                      <td className="text-sm font-medium">{log.programs_processed}</td>
                      <td className="text-sm text-green-600 dark:text-green-400">
                        +{log.programs_created}
                      </td>
                      <td className="text-sm text-blue-600 dark:text-blue-400">
                        {log.programs_updated}
                      </td>
                      <td className="text-sm text-red-600 dark:text-red-400">
                        {log.programs_failed}
                      </td>
                      <td className="text-sm">
                        {log.duration_seconds ? `${log.duration_seconds}s` : '-'}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Create Program Modal */}
      {showProgramModal && (
        <div className="modal-overlay">
          <div className="modal-content max-w-2xl">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
              Add EPG Program
            </h2>

            <form className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">Stream ID *</label>
                  <input
                    type="number"
                    className="input"
                    value={programFormData.stream_id || ''}
                    onChange={(e) =>
                      setProgramFormData({ ...programFormData, stream_id: parseInt(e.target.value) })
                    }
                    required
                  />
                </div>

                <div>
                  <label className="label">Category</label>
                  <input
                    type="text"
                    className="input"
                    value={programFormData.category}
                    onChange={(e) =>
                      setProgramFormData({ ...programFormData, category: e.target.value })
                    }
                    placeholder="Drama, News, Sports..."
                  />
                </div>
              </div>

              <div>
                <label className="label">Title *</label>
                <input
                  type="text"
                  className="input"
                  value={programFormData.title}
                  onChange={(e) =>
                    setProgramFormData({ ...programFormData, title: e.target.value })
                  }
                  required
                />
              </div>

              <div>
                <label className="label">Description</label>
                <textarea
                  className="input"
                  rows={3}
                  value={programFormData.description}
                  onChange={(e) =>
                    setProgramFormData({ ...programFormData, description: e.target.value })
                  }
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">Start Time *</label>
                  <input
                    type="datetime-local"
                    className="input"
                    value={programFormData.start_time}
                    onChange={(e) =>
                      setProgramFormData({ ...programFormData, start_time: e.target.value })
                    }
                    required
                  />
                </div>

                <div>
                  <label className="label">End Time *</label>
                  <input
                    type="datetime-local"
                    className="input"
                    value={programFormData.end_time}
                    onChange={(e) =>
                      setProgramFormData({ ...programFormData, end_time: e.target.value })
                    }
                    required
                  />
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="label">Season</label>
                  <input
                    type="number"
                    className="input"
                    value={programFormData.season_number || ''}
                    onChange={(e) =>
                      setProgramFormData({
                        ...programFormData,
                        season_number: parseInt(e.target.value) || undefined,
                      })
                    }
                  />
                </div>

                <div>
                  <label className="label">Episode</label>
                  <input
                    type="number"
                    className="input"
                    value={programFormData.episode_number || ''}
                    onChange={(e) =>
                      setProgramFormData({
                        ...programFormData,
                        episode_number: parseInt(e.target.value) || undefined,
                      })
                    }
                  />
                </div>

                <div>
                  <label className="label">Year</label>
                  <input
                    type="number"
                    className="input"
                    value={programFormData.year || ''}
                    onChange={(e) =>
                      setProgramFormData({
                        ...programFormData,
                        year: parseInt(e.target.value) || undefined,
                      })
                    }
                  />
                </div>
              </div>

              <div className="flex items-center space-x-6">
                <label className="flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    className="mr-2"
                    checked={programFormData.is_live}
                    onChange={(e) =>
                      setProgramFormData({ ...programFormData, is_live: e.target.checked })
                    }
                  />
                  <span className="text-sm font-medium">Live</span>
                </label>

                <label className="flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    className="mr-2"
                    checked={programFormData.is_repeat}
                    onChange={(e) =>
                      setProgramFormData({ ...programFormData, is_repeat: e.target.checked })
                    }
                  />
                  <span className="text-sm font-medium">Repeat</span>
                </label>

                <label className="flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    className="mr-2"
                    checked={programFormData.is_premiere}
                    onChange={(e) =>
                      setProgramFormData({ ...programFormData, is_premiere: e.target.checked })
                    }
                  />
                  <span className="text-sm font-medium">Premiere</span>
                </label>
              </div>

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowProgramModal(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button type="button" onClick={handleCreateProgram} className="btn-primary">
                  Create Program
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Create Source Modal */}
      {showSourceModal && (
        <div className="modal-overlay">
          <div className="modal-content max-w-lg">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
              Add EPG Source
            </h2>

            <form className="space-y-4">
              <div>
                <label className="label">Source Name *</label>
                <input
                  type="text"
                  className="input"
                  value={sourceFormData.name}
                  onChange={(e) => setSourceFormData({ ...sourceFormData, name: e.target.value })}
                  required
                  placeholder="Main XMLTV Source"
                />
              </div>

              <div>
                <label className="label">Source Type *</label>
                <select
                  className="input"
                  value={sourceFormData.source_type}
                  onChange={(e) =>
                    setSourceFormData({
                      ...sourceFormData,
                      source_type: e.target.value as 'xmltv' | 'json' | 'api' | 'manual',
                    })
                  }
                  required
                >
                  <option value="xmltv">XMLTV</option>
                  <option value="json">JSON</option>
                  <option value="api">API</option>
                  <option value="manual">Manual</option>
                </select>
              </div>

              <div>
                <label className="label">Source URL *</label>
                <input
                  type="url"
                  className="input"
                  value={sourceFormData.url}
                  onChange={(e) => setSourceFormData({ ...sourceFormData, url: e.target.value })}
                  required
                  placeholder="https://example.com/epg.xml"
                />
              </div>

              <div>
                <label className="label">Update Interval (seconds)</label>
                <input
                  type="number"
                  className="input"
                  value={sourceFormData.update_interval}
                  onChange={(e) =>
                    setSourceFormData({
                      ...sourceFormData,
                      update_interval: parseInt(e.target.value),
                    })
                  }
                  min="60"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Current: {sourceFormData.update_interval / 60} minutes
                </p>
              </div>

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowSourceModal(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button type="button" onClick={handleCreateSource} className="btn-primary">
                  Create Source
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* View Program Modal */}
      {showViewModal && selectedProgram && (
        <div className="modal-overlay">
          <div className="modal-content max-w-2xl">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
              {selectedProgram.title}
            </h2>

            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Stream</p>
                  <p className="font-medium">{selectedProgram.stream_name}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Category</p>
                  <p className="font-medium">{selectedProgram.category}</p>
                </div>
              </div>

              {selectedProgram.description && (
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Description</p>
                  <p className="text-gray-900 dark:text-white">{selectedProgram.description}</p>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">Start Time</p>
                  <p className="font-medium">{formatDateTime(selectedProgram.start_time)}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400">End Time</p>
                  <p className="font-medium">{formatDateTime(selectedProgram.end_time)}</p>
                </div>
              </div>

              {(selectedProgram.season_number || selectedProgram.episode_number) && (
                <div className="grid grid-cols-2 gap-4">
                  {selectedProgram.season_number && (
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Season</p>
                      <p className="font-medium">{selectedProgram.season_number}</p>
                    </div>
                  )}
                  {selectedProgram.episode_number && (
                    <div>
                      <p className="text-sm text-gray-500 dark:text-gray-400">Episode</p>
                      <p className="font-medium">{selectedProgram.episode_number}</p>
                    </div>
                  )}
                </div>
              )}

              {selectedProgram.directors && selectedProgram.directors.length > 0 && (
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">Directors</p>
                  <div className="flex flex-wrap gap-2">
                    {selectedProgram.directors.map((director, i) => (
                      <span
                        key={i}
                        className="px-3 py-1 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 rounded-full text-sm"
                      >
                        {director}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {selectedProgram.actors && selectedProgram.actors.length > 0 && (
                <div>
                  <p className="text-sm text-gray-500 dark:text-gray-400 mb-2">Actors</p>
                  <div className="flex flex-wrap gap-2">
                    {selectedProgram.actors.map((actor, i) => (
                      <span
                        key={i}
                        className="px-3 py-1 bg-gray-100 dark:bg-dark-800 text-gray-700 dark:text-gray-300 rounded-full text-sm"
                      >
                        {actor}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              <div className="flex justify-end">
                <button onClick={() => setShowViewModal(false)} className="btn-secondary">
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
