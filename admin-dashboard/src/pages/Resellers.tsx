import { useState, useEffect } from 'react';
import { MagnifyingGlassIcon, PlusIcon, PencilIcon, TrashIcon, BanknotesIcon, UsersIcon, ChartBarIcon } from '@heroicons/react/24/outline';
import axios from 'axios';
import toast from 'react-hot-toast';

interface Reseller {
  id: number;
  user_id: number;
  username: string;
  email: string;
  parent_id?: number;
  parent_username?: string;
  credits: number;
  commission_rate: number;
  can_create_resellers: boolean;
  max_users: number;
  max_resellers: number;
  current_users: number;
  current_resellers: number;
  is_active: boolean;
  notes?: string;
  created_at: string;
}

interface ResellerFormData {
  user_id: number;
  parent_id?: number;
  credits: number;
  commission_rate: number;
  can_create_resellers: boolean;
  max_users: number;
  max_resellers: number;
  notes: string;
}

export default function Resellers() {
  const [resellers, setResellers] = useState<Reseller[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showCreditsModal, setShowCreditsModal] = useState(false);
  const [selectedReseller, setSelectedReseller] = useState<Reseller | null>(null);
  const [formData, setFormData] = useState<ResellerFormData>({
    user_id: 0,
    credits: 0,
    commission_rate: 10,
    can_create_resellers: false,
    max_users: 100,
    max_resellers: 0,
    notes: '',
  });
  const [creditsAmount, setCreditsAmount] = useState(0);
  const [creditsDescription, setCreditsDescription] = useState('');

  useEffect(() => {
    loadResellers();
  }, []);

  const loadResellers = async () => {
    try {
      setLoading(true);
      // Mock data for now
      setTimeout(() => {
        setResellers([
          {
            id: 1,
            user_id: 10,
            username: 'reseller1',
            email: 'reseller1@iptv.com',
            credits: 1500.00,
            commission_rate: 15,
            can_create_resellers: true,
            max_users: 200,
            max_resellers: 10,
            current_users: 45,
            current_resellers: 3,
            is_active: true,
            created_at: '2024-01-15',
          },
          {
            id: 2,
            user_id: 11,
            username: 'reseller2',
            email: 'reseller2@iptv.com',
            parent_id: 1,
            parent_username: 'reseller1',
            credits: 750.00,
            commission_rate: 10,
            can_create_resellers: false,
            max_users: 50,
            max_resellers: 0,
            current_users: 12,
            current_resellers: 0,
            is_active: true,
            created_at: '2024-02-20',
          },
        ]);
        setLoading(false);
      }, 500);
    } catch (error) {
      console.error('Failed to load resellers:', error);
      toast.error('Failed to load resellers');
      setLoading(false);
    }
  };

  const handleCreateReseller = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      // await axios.post('/api/v1/admin/resellers', formData);
      toast.success('Reseller created successfully');
      setShowCreateModal(false);
      loadResellers();
      resetForm();
    } catch (error) {
      toast.error('Failed to create reseller');
    }
  };

  const handleUpdateReseller = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedReseller) return;

    try {
      // await axios.put(`/api/v1/admin/resellers/${selectedReseller.id}`, formData);
      toast.success('Reseller updated successfully');
      setShowEditModal(false);
      loadResellers();
      resetForm();
    } catch (error) {
      toast.error('Failed to update reseller');
    }
  };

  const handleAddCredits = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedReseller) return;

    try {
      // await axios.post(`/api/v1/admin/resellers/${selectedReseller.id}/credits`, {
      //   amount: creditsAmount,
      //   description: creditsDescription,
      // });
      toast.success(`${creditsAmount} credits added successfully`);
      setShowCreditsModal(false);
      loadResellers();
      setCreditsAmount(0);
      setCreditsDescription('');
    } catch (error) {
      toast.error('Failed to add credits');
    }
  };

  const handleDeleteReseller = async (reseller: Reseller) => {
    if (!confirm(`Are you sure you want to delete reseller "${reseller.username}"?`)) return;

    try {
      // await axios.delete(`/api/v1/admin/resellers/${reseller.id}`);
      toast.success('Reseller deleted successfully');
      loadResellers();
    } catch (error) {
      toast.error('Failed to delete reseller');
    }
  };

  const openEditModal = (reseller: Reseller) => {
    setSelectedReseller(reseller);
    setFormData({
      user_id: reseller.user_id,
      parent_id: reseller.parent_id,
      credits: reseller.credits,
      commission_rate: reseller.commission_rate,
      can_create_resellers: reseller.can_create_resellers,
      max_users: reseller.max_users,
      max_resellers: reseller.max_resellers,
      notes: reseller.notes || '',
    });
    setShowEditModal(true);
  };

  const openCreditsModal = (reseller: Reseller) => {
    setSelectedReseller(reseller);
    setShowCreditsModal(true);
  };

  const resetForm = () => {
    setFormData({
      user_id: 0,
      credits: 0,
      commission_rate: 10,
      can_create_resellers: false,
      max_users: 100,
      max_resellers: 0,
      notes: '',
    });
    setSelectedReseller(null);
  };

  const filteredResellers = resellers.filter(reseller =>
    reseller.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    reseller.email.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Reseller Management</h2>
          <p className="text-gray-600 dark:text-gray-400 mt-1">Manage resellers, credits, and commissions</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="btn-primary flex items-center space-x-2"
        >
          <PlusIcon className="h-5 w-5" />
          <span>Add Reseller</span>
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="stat-card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Total Resellers</p>
              <p className="text-3xl font-bold mt-2">{resellers.length}</p>
            </div>
            <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-xl">
              <UsersIcon className="h-8 w-8 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Total Credits</p>
              <p className="text-3xl font-bold mt-2">
                ${resellers.reduce((sum, r) => sum + r.credits, 0).toFixed(2)}
              </p>
            </div>
            <div className="p-3 bg-green-100 dark:bg-green-900/30 rounded-xl">
              <BanknotesIcon className="h-8 w-8 text-green-600 dark:text-green-400" />
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Total Customers</p>
              <p className="text-3xl font-bold mt-2">
                {resellers.reduce((sum, r) => sum + r.current_users, 0)}
              </p>
            </div>
            <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-xl">
              <ChartBarIcon className="h-8 w-8 text-purple-600 dark:text-purple-400" />
            </div>
          </div>
        </div>
      </div>

      {/* Search */}
      <div className="card p-4">
        <div className="relative">
          <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search resellers..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="input pl-10"
          />
        </div>
      </div>

      {/* Resellers Table */}
      <div className="card overflow-hidden">
        <div className="table-container">
          <table className="table">
            <thead>
              <tr>
                <th>Reseller</th>
                <th>Parent</th>
                <th>Credits</th>
                <th>Commission</th>
                <th>Users</th>
                <th>Sub-Resellers</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array(3).fill(0).map((_, i) => (
                  <tr key={i}>
                    <td colSpan={8}>
                      <div className="h-12 skeleton rounded" />
                    </td>
                  </tr>
                ))
              ) : (
                filteredResellers.map(reseller => (
                  <tr key={reseller.id}>
                    <td>
                      <div>
                        <div className="font-medium text-gray-900 dark:text-white">{reseller.username}</div>
                        <div className="text-sm text-gray-500">{reseller.email}</div>
                      </div>
                    </td>
                    <td>
                      {reseller.parent_username ? (
                        <span className="text-sm">{reseller.parent_username}</span>
                      ) : (
                        <span className="text-sm text-gray-400">-</span>
                      )}
                    </td>
                    <td>
                      <span className="font-semibold text-green-600 dark:text-green-400">
                        ${reseller.credits.toFixed(2)}
                      </span>
                    </td>
                    <td>{reseller.commission_rate}%</td>
                    <td>
                      <span className="text-sm">
                        {reseller.current_users} / {reseller.max_users}
                      </span>
                    </td>
                    <td>
                      <span className="text-sm">
                        {reseller.current_resellers} / {reseller.max_resellers}
                      </span>
                    </td>
                    <td>
                      <span className={`badge ${reseller.is_active ? 'badge-success' : 'badge-danger'}`}>
                        {reseller.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => openCreditsModal(reseller)}
                          className="p-2 text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20 rounded-lg transition-colors"
                          title="Add Credits"
                        >
                          <BanknotesIcon className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => openEditModal(reseller)}
                          className="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
                          title="Edit"
                        >
                          <PencilIcon className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteReseller(reseller)}
                          className="p-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
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

      {/* Create Reseller Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4">
            <div className="fixed inset-0 bg-gray-600/75 transition-opacity" onClick={() => setShowCreateModal(false)} />

            <div className="relative bg-white dark:bg-dark-900 rounded-2xl shadow-xl max-w-2xl w-full p-6">
              <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-6">Create New Reseller</h3>

              <form onSubmit={handleCreateReseller} className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      User ID
                    </label>
                    <input
                      type="number"
                      required
                      className="input"
                      value={formData.user_id || ''}
                      onChange={(e) => setFormData({ ...formData, user_id: parseInt(e.target.value) })}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Parent Reseller ID (Optional)
                    </label>
                    <input
                      type="number"
                      className="input"
                      value={formData.parent_id || ''}
                      onChange={(e) => setFormData({ ...formData, parent_id: parseInt(e.target.value) || undefined })}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Initial Credits
                    </label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      required
                      className="input"
                      value={formData.credits}
                      onChange={(e) => setFormData({ ...formData, credits: parseFloat(e.target.value) })}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Commission Rate (%)
                    </label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      max="100"
                      required
                      className="input"
                      value={formData.commission_rate}
                      onChange={(e) => setFormData({ ...formData, commission_rate: parseFloat(e.target.value) })}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Max Users
                    </label>
                    <input
                      type="number"
                      min="0"
                      required
                      className="input"
                      value={formData.max_users}
                      onChange={(e) => setFormData({ ...formData, max_users: parseInt(e.target.value) })}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Max Sub-Resellers
                    </label>
                    <input
                      type="number"
                      min="0"
                      required
                      className="input"
                      value={formData.max_resellers}
                      onChange={(e) => setFormData({ ...formData, max_resellers: parseInt(e.target.value) })}
                    />
                  </div>
                </div>

                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={formData.can_create_resellers}
                      onChange={(e) => setFormData({ ...formData, can_create_resellers: e.target.checked })}
                      className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      Can create sub-resellers
                    </span>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Notes (Optional)
                  </label>
                  <textarea
                    rows={3}
                    className="input"
                    value={formData.notes}
                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  />
                </div>

                <div className="flex justify-end space-x-3 mt-6">
                  <button
                    type="button"
                    onClick={() => { setShowCreateModal(false); resetForm(); }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button type="submit" className="btn-primary">
                    Create Reseller
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Add Credits Modal */}
      {showCreditsModal && selectedReseller && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4">
            <div className="fixed inset-0 bg-gray-600/75 transition-opacity" onClick={() => setShowCreditsModal(false)} />

            <div className="relative bg-white dark:bg-dark-900 rounded-2xl shadow-xl max-w-md w-full p-6">
              <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-6">
                Add Credits to {selectedReseller.username}
              </h3>

              <div className="mb-6 p-4 bg-gray-50 dark:bg-dark-800 rounded-lg">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600 dark:text-gray-400">Current Balance:</span>
                  <span className="text-lg font-bold text-green-600 dark:text-green-400">
                    ${selectedReseller.credits.toFixed(2)}
                  </span>
                </div>
              </div>

              <form onSubmit={handleAddCredits} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Amount
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    required
                    className="input"
                    value={creditsAmount || ''}
                    onChange={(e) => setCreditsAmount(parseFloat(e.target.value))}
                    placeholder="Enter amount to add"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Description
                  </label>
                  <input
                    type="text"
                    required
                    className="input"
                    value={creditsDescription}
                    onChange={(e) => setCreditsDescription(e.target.value)}
                    placeholder="e.g., Monthly credit top-up"
                  />
                </div>

                {creditsAmount > 0 && (
                  <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-gray-600 dark:text-gray-400">New Balance:</span>
                      <span className="text-lg font-bold text-blue-600 dark:text-blue-400">
                        ${(selectedReseller.credits + creditsAmount).toFixed(2)}
                      </span>
                    </div>
                  </div>
                )}

                <div className="flex justify-end space-x-3 mt-6">
                  <button
                    type="button"
                    onClick={() => { setShowCreditsModal(false); setCreditsAmount(0); setCreditsDescription(''); }}
                    className="btn-secondary"
                  >
                    Cancel
                  </button>
                  <button type="submit" className="btn-primary">
                    Add Credits
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
