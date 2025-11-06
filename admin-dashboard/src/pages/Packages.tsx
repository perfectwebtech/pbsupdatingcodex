import React, { useState, useEffect } from 'react';
import {
  CubeIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  TrashIcon,
  PencilIcon,
  CheckCircleIcon,
  XMarkIcon,
  StarIcon,
} from '@heroicons/react/24/outline';
import { StarIcon as StarIconSolid } from '@heroicons/react/24/solid';

interface Package {
  id: number;
  name: string;
  description?: string;
  price: number;
  currency: string;
  billing_cycle: 'monthly' | 'quarterly' | 'yearly' | 'lifetime';
  max_connections: number;
  stream_count: number;
  is_active: boolean;
  is_featured: boolean;
  subscriber_count: number;
  features: string[];
  created_at: string;
  updated_at: string;
}

interface PackageStats {
  total_packages: number;
  active_packages: number;
  total_subscribers: number;
  total_revenue: number;
}

interface PackageFormData {
  name: string;
  description?: string;
  price: number;
  currency: string;
  billing_cycle: 'monthly' | 'quarterly' | 'yearly' | 'lifetime';
  max_connections: number;
  is_active: boolean;
  is_featured: boolean;
  features: string[];
}

const Packages: React.FC = () => {
  const [packages, setPackages] = useState<Package[]>([]);
  const [stats, setStats] = useState<PackageStats>({
    total_packages: 0,
    active_packages: 0,
    total_subscribers: 0,
    total_revenue: 0,
  });

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedPackage, setSelectedPackage] = useState<Package | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterCycle, setFilterCycle] = useState<string>('all');

  const [formData, setFormData] = useState<PackageFormData>({
    name: '',
    description: '',
    price: 0,
    currency: 'USD',
    billing_cycle: 'monthly',
    max_connections: 1,
    is_active: true,
    is_featured: false,
    features: [],
  });

  const [featureInput, setFeatureInput] = useState('');

  useEffect(() => {
    loadMockData();
  }, []);

  const loadMockData = () => {
    const mockPackages: Package[] = [
      {
        id: 1,
        name: 'Basic Plan',
        description: 'Perfect for individuals',
        price: 9.99,
        currency: 'USD',
        billing_cycle: 'monthly',
        max_connections: 1,
        stream_count: 1500,
        is_active: true,
        is_featured: false,
        subscriber_count: 2450,
        features: ['1 Connection', '1500+ Channels', 'SD/HD Quality', 'VOD Library', '24/7 Support'],
        created_at: new Date(Date.now() - 180 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 1 * 86400000).toISOString(),
      },
      {
        id: 2,
        name: 'Standard Plan',
        description: 'Great for families',
        price: 14.99,
        currency: 'USD',
        billing_cycle: 'monthly',
        max_connections: 2,
        stream_count: 2500,
        is_active: true,
        is_featured: true,
        subscriber_count: 4890,
        features: ['2 Connections', '2500+ Channels', 'Full HD Quality', 'VOD & Series', 'Catch-up TV', 'Priority Support'],
        created_at: new Date(Date.now() - 170 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 2 * 86400000).toISOString(),
      },
      {
        id: 3,
        name: 'Premium Plan',
        description: 'Ultimate entertainment',
        price: 19.99,
        currency: 'USD',
        billing_cycle: 'monthly',
        max_connections: 3,
        stream_count: 3500,
        is_active: true,
        is_featured: true,
        subscriber_count: 3210,
        features: ['3 Connections', '3500+ Channels', '4K Quality', 'VOD & Series', 'Catch-up TV', 'EPG', 'DVR Recording', 'VIP Support'],
        created_at: new Date(Date.now() - 160 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 3 * 86400000).toISOString(),
      },
      {
        id: 4,
        name: 'Annual Premium',
        description: 'Save 20% with annual billing',
        price: 199.99,
        currency: 'USD',
        billing_cycle: 'yearly',
        max_connections: 3,
        stream_count: 3500,
        is_active: true,
        is_featured: false,
        subscriber_count: 1560,
        features: ['3 Connections', '3500+ Channels', '4K Quality', 'All Premium Features', '20% Discount', 'Free Setup'],
        created_at: new Date(Date.now() - 150 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 4 * 86400000).toISOString(),
      },
      {
        id: 5,
        name: 'Lifetime Access',
        description: 'One-time payment, lifetime access',
        price: 499.99,
        currency: 'USD',
        billing_cycle: 'lifetime',
        max_connections: 5,
        stream_count: 5000,
        is_active: true,
        is_featured: true,
        subscriber_count: 890,
        features: ['5 Connections', '5000+ Channels', '4K Quality', 'Lifetime Updates', 'All Features', 'VIP Support', 'Free Migrations'],
        created_at: new Date(Date.now() - 140 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 5 * 86400000).toISOString(),
      },
    ];

    const mockStats: PackageStats = {
      total_packages: mockPackages.length,
      active_packages: mockPackages.filter(p => p.is_active).length,
      total_subscribers: mockPackages.reduce((sum, p) => sum + p.subscriber_count, 0),
      total_revenue: mockPackages.reduce((sum, p) => {
        if (p.billing_cycle === 'monthly') return sum + (p.price * p.subscriber_count);
        if (p.billing_cycle === 'yearly') return sum + ((p.price / 12) * p.subscriber_count);
        return sum;
      }, 0),
    };

    setPackages(mockPackages);
    setStats(mockStats);
  };

  const handleCreatePackage = async () => {
    if (!formData.name || formData.price <= 0) {
      alert('Please fill in all required fields');
      return;
    }

    console.log('Creating package:', formData);

    const newPackage: Package = {
      id: packages.length + 1,
      ...formData,
      stream_count: 1000,
      subscriber_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setPackages([newPackage, ...packages]);

    setShowCreateModal(false);
    resetForm();
    alert('Package created successfully!');
  };

  const handleEditPackage = async () => {
    if (!selectedPackage || !formData.name) {
      alert('Please fill in all required fields');
      return;
    }

    console.log('Updating package:', selectedPackage.id, formData);

    setPackages(packages.map(p =>
      p.id === selectedPackage.id
        ? { ...p, ...formData, updated_at: new Date().toISOString() }
        : p
    ));

    setShowEditModal(false);
    setSelectedPackage(null);
    resetForm();
    alert('Package updated successfully!');
  };

  const handleDeletePackage = async (pkg: Package) => {
    if (pkg.subscriber_count > 0) {
      alert(`Cannot delete package with ${pkg.subscriber_count} active subscribers. Please migrate subscribers first.`);
      return;
    }

    if (!confirm(`Delete package "${pkg.name}"?`)) return;

    console.log('Deleting package:', pkg.id);
    setPackages(packages.filter(p => p.id !== pkg.id));
    alert('Package deleted successfully!');
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      price: 0,
      currency: 'USD',
      billing_cycle: 'monthly',
      max_connections: 1,
      is_active: true,
      is_featured: false,
      features: [],
    });
    setFeatureInput('');
  };

  const openEditModal = (pkg: Package) => {
    setSelectedPackage(pkg);
    setFormData({
      name: pkg.name,
      description: pkg.description,
      price: pkg.price,
      currency: pkg.currency,
      billing_cycle: pkg.billing_cycle,
      max_connections: pkg.max_connections,
      is_active: pkg.is_active,
      is_featured: pkg.is_featured,
      features: [...pkg.features],
    });
    setShowEditModal(true);
  };

  const addFeature = () => {
    if (featureInput.trim() && !formData.features.includes(featureInput.trim())) {
      setFormData({ ...formData, features: [...formData.features, featureInput.trim()] });
      setFeatureInput('');
    }
  };

  const removeFeature = (feature: string) => {
    setFormData({ ...formData, features: formData.features.filter(f => f !== feature) });
  };

  const getBillingCycleName = (cycle: string): string => {
    const names: Record<string, string> = {
      monthly: 'Monthly',
      quarterly: 'Quarterly',
      yearly: 'Yearly',
      lifetime: 'Lifetime',
    };
    return names[cycle] || cycle;
  };

  const filteredPackages = packages.filter(pkg => {
    const matchesSearch = pkg.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pkg.description?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesCycle = filterCycle === 'all' || pkg.billing_cycle === filterCycle;
    return matchesSearch && matchesCycle;
  });

  return (
    <div className="p-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <CubeIcon className="w-8 h-8 text-blue-500" />
          Package Management
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          Manage subscription packages and pricing plans
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Total Packages</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-2">{stats.total_packages}</p>
            </div>
            <CubeIcon className="w-12 h-12 text-blue-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Active Packages</p>
              <p className="text-3xl font-bold text-green-600 dark:text-green-400 mt-2">{stats.active_packages}</p>
            </div>
            <CheckCircleIcon className="w-12 h-12 text-green-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Total Subscribers</p>
              <p className="text-3xl font-bold text-purple-600 dark:text-purple-400 mt-2">{stats.total_subscribers.toLocaleString()}</p>
            </div>
            <div className="text-5xl">👥</div>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Monthly Revenue</p>
              <p className="text-3xl font-bold text-orange-600 dark:text-orange-400 mt-2">${stats.total_revenue.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</p>
            </div>
            <div className="text-5xl">💰</div>
          </div>
        </div>
      </div>

      {/* Toolbar */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <div className="flex flex-col lg:flex-row gap-4 items-center justify-between">
          <div className="flex flex-col sm:flex-row gap-4 w-full lg:w-auto">
            <div className="relative flex-1 lg:w-64">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search packages..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              />
            </div>

            <select
              value={filterCycle}
              onChange={(e) => setFilterCycle(e.target.value)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="all">All Billing Cycles</option>
              <option value="monthly">Monthly</option>
              <option value="quarterly">Quarterly</option>
              <option value="yearly">Yearly</option>
              <option value="lifetime">Lifetime</option>
            </select>
          </div>

          <button
            onClick={() => setShowCreateModal(true)}
            className="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg flex items-center gap-2 justify-center transition-colors"
          >
            <PlusIcon className="w-5 h-5" />
            Create Package
          </button>
        </div>
      </div>

      {/* Packages Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredPackages.map((pkg) => (
          <div
            key={pkg.id}
            className={`bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden hover:shadow-xl transition-shadow ${
              pkg.is_featured ? 'ring-2 ring-blue-500' : ''
            }`}
          >
            {/* Header */}
            <div className={`p-6 ${pkg.is_featured ? 'bg-gradient-to-r from-blue-600 to-purple-600' : 'bg-gray-50 dark:bg-gray-700'}`}>
              <div className="flex items-start justify-between">
                <div>
                  <h3 className={`text-2xl font-bold ${pkg.is_featured ? 'text-white' : 'text-gray-900 dark:text-white'}`}>
                    {pkg.name}
                  </h3>
                  <p className={`text-sm mt-1 ${pkg.is_featured ? 'text-blue-100' : 'text-gray-600 dark:text-gray-400'}`}>
                    {pkg.description}
                  </p>
                </div>
                {pkg.is_featured && (
                  <StarIconSolid className="w-6 h-6 text-yellow-300" />
                )}
              </div>

              <div className="mt-4">
                <span className={`text-4xl font-bold ${pkg.is_featured ? 'text-white' : 'text-gray-900 dark:text-white'}`}>
                  ${pkg.price}
                </span>
                <span className={`text-sm ml-2 ${pkg.is_featured ? 'text-blue-100' : 'text-gray-600 dark:text-gray-400'}`}>
                  /{getBillingCycleName(pkg.billing_cycle).toLowerCase()}
                </span>
              </div>
            </div>

            {/* Features */}
            <div className="p-6">
              <ul className="space-y-3 mb-6">
                {pkg.features.map((feature, index) => (
                  <li key={index} className="flex items-start gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <CheckCircleIcon className="w-5 h-5 text-green-500 flex-shrink-0 mt-0.5" />
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>

              <div className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                <div className="flex justify-between mb-2">
                  <span>Active Subscribers:</span>
                  <span className="font-semibold text-gray-900 dark:text-white">{pkg.subscriber_count.toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span>Status:</span>
                  <span className={`font-semibold ${pkg.is_active ? 'text-green-600' : 'text-red-600'}`}>
                    {pkg.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-2 pt-4 border-t border-gray-200 dark:border-gray-700">
                <button
                  onClick={() => openEditModal(pkg)}
                  className="flex-1 px-4 py-2 bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200 rounded-lg hover:bg-blue-200 dark:hover:bg-blue-800 transition-colors flex items-center justify-center gap-2"
                >
                  <PencilIcon className="w-4 h-4" />
                  Edit
                </button>
                <button
                  onClick={() => handleDeletePackage(pkg)}
                  className="flex-1 px-4 py-2 bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-200 rounded-lg hover:bg-red-200 dark:hover:bg-red-800 transition-colors flex items-center justify-center gap-2"
                >
                  <TrashIcon className="w-4 h-4" />
                  Delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Empty State */}
      {filteredPackages.length === 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-12 text-center">
          <CubeIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-500 dark:text-gray-400 text-lg">No packages found</p>
        </div>
      )}

      {/* Create/Edit Modal */}
      {(showCreateModal || showEditModal) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
                {showCreateModal ? 'Create New Package' : 'Edit Package'}
              </h2>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Package Name *
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="e.g., Premium Plan"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Description
                  </label>
                  <textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    rows={2}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="Package description..."
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Price *
                    </label>
                    <input
                      type="number"
                      step="0.01"
                      value={formData.price}
                      onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="9.99"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Billing Cycle
                    </label>
                    <select
                      value={formData.billing_cycle}
                      onChange={(e) => setFormData({ ...formData, billing_cycle: e.target.value as any })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    >
                      <option value="monthly">Monthly</option>
                      <option value="quarterly">Quarterly</option>
                      <option value="yearly">Yearly</option>
                      <option value="lifetime">Lifetime</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Max Connections
                    </label>
                    <input
                      type="number"
                      value={formData.max_connections}
                      onChange={(e) => setFormData({ ...formData, max_connections: parseInt(e.target.value) })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="1"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Features
                  </label>
                  <div className="flex gap-2 mb-2">
                    <input
                      type="text"
                      value={featureInput}
                      onChange={(e) => setFeatureInput(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addFeature())}
                      className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="Enter feature and press Enter"
                    />
                    <button
                      onClick={addFeature}
                      className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                    >
                      Add
                    </button>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {formData.features.map((feature, index) => (
                      <span
                        key={index}
                        className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded-full text-sm"
                      >
                        {feature}
                        <button
                          onClick={() => removeFeature(feature)}
                          className="hover:text-blue-600"
                        >
                          <XMarkIcon className="w-4 h-4" />
                        </button>
                      </span>
                    ))}
                  </div>
                </div>

                <div className="flex items-center gap-6">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formData.is_active}
                      onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded"
                    />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Active</span>
                  </label>

                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formData.is_featured}
                      onChange={(e) => setFormData({ ...formData, is_featured: e.target.checked })}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded"
                    />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Featured</span>
                  </label>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={showCreateModal ? handleCreatePackage : handleEditPackage}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  {showCreateModal ? 'Create Package' : 'Update Package'}
                </button>
                <button
                  onClick={() => {
                    setShowCreateModal(false);
                    setShowEditModal(false);
                    setSelectedPackage(null);
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
    </div>
  );
};

export default Packages;
