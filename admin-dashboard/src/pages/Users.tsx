import { useState, useEffect } from 'react';
import { MagnifyingGlassIcon, PlusIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline';

export default function Users() {
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    // Mock data
    setTimeout(() => {
      setUsers([
        { id: 1, username: 'admin', email: 'admin@iptv.com', package: 'Enterprise', status: 'active', connections: '2/10', created: '2024-01-15' },
        { id: 2, username: 'testuser', email: 'test@iptv.com', package: 'Standard', status: 'active', connections: '1/2', created: '2024-02-20' },
        { id: 3, username: 'premium_user', email: 'premium@iptv.com', package: 'Premium', status: 'active', connections: '3/5', created: '2024-03-10' },
        { id: 4, username: 'trial_user', email: 'trial@iptv.com', package: 'Free Trial', status: 'expired', connections: '0/1', created: '2024-04-01' },
      ]);
      setLoading(false);
    }, 500);
  }, []);

  const filteredUsers = users.filter(user =>
    user.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    user.email.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Users Management</h2>
          <p className="text-gray-600 dark:text-gray-400 mt-1">Manage user accounts and subscriptions</p>
        </div>
        <button className="btn-primary flex items-center space-x-2">
          <PlusIcon className="h-5 w-5" />
          <span>Add User</span>
        </button>
      </div>

      {/* Search and Filters */}
      <div className="card p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1 relative">
            <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Search users..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="input pl-10"
            />
          </div>
          <select className="input w-full sm:w-48">
            <option>All Packages</option>
            <option>Enterprise</option>
            <option>Premium</option>
            <option>Standard</option>
            <option>Free Trial</option>
          </select>
          <select className="input w-full sm:w-48">
            <option>All Status</option>
            <option>Active</option>
            <option>Expired</option>
            <option>Suspended</option>
          </select>
        </div>
      </div>

      {/* Users Table */}
      <div className="card overflow-hidden">
        <div className="table-container">
          <table className="table">
            <thead>
              <tr>
                <th>User</th>
                <th>Package</th>
                <th>Connections</th>
                <th>Status</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array(5).fill(0).map((_, i) => (
                  <tr key={i}>
                    <td colSpan={6}>
                      <div className="h-12 skeleton rounded" />
                    </td>
                  </tr>
                ))
              ) : (
                filteredUsers.map(user => (
                  <tr key={user.id}>
                    <td>
                      <div>
                        <div className="font-medium text-gray-900 dark:text-white">{user.username}</div>
                        <div className="text-sm text-gray-500">{user.email}</div>
                      </div>
                    </td>
                    <td>
                      <span className="badge badge-info">{user.package}</span>
                    </td>
                    <td>{user.connections}</td>
                    <td>
                      <span className={`badge ${user.status === 'active' ? 'badge-success' : 'badge-danger'}`}>
                        {user.status}
                      </span>
                    </td>
                    <td>{user.created}</td>
                    <td>
                      <div className="flex space-x-2">
                        <button className="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors">
                          <PencilIcon className="h-4 w-4" />
                        </button>
                        <button className="p-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors">
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
    </div>
  );
}
