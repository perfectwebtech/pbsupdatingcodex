import { useEffect, useState } from 'react';
import {
  CreditCardIcon,
  DocumentTextIcon,
  BanknotesIcon,
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
  ArrowPathIcon,
  FunnelIcon,
  MagnifyingGlassIcon,
  PlusIcon,
  EyeIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import { billingAPI } from '../services/api';

interface Invoice {
  id: number;
  invoice_number: string;
  user_id: number;
  username: string;
  email: string;
  reseller_id?: number;
  package_id?: number;
  package_name?: string;
  description: string;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  status: 'pending' | 'paid' | 'overdue' | 'cancelled' | 'refunded';
  issue_date: string;
  due_date?: string;
  paid_at?: string;
  payment_method?: string;
  payment_gateway?: string;
  transaction_id?: string;
  notes?: string;
  created_at: string;
}

interface InvoiceFormData {
  user_id: number;
  package_id?: number;
  description: string;
  subtotal: number;
  tax: number;
  discount: number;
  due_date?: string;
  notes: string;
}

interface PaymentFormData {
  invoice_id: number;
  payment_gateway: 'stripe' | 'paypal' | 'crypto' | 'bank_transfer';
  amount: number;
  payment_details: Record<string, any>;
}

interface RevenueStats {
  total_revenue: number;
  monthly_revenue: number;
  weekly_revenue: number;
  daily_revenue: number;
  pending_amount: number;
  overdue_amount: number;
  total_invoices: number;
  paid_invoices: number;
  pending_invoices: number;
  overdue_invoices: number;
}

export default function Billing() {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [stats, setStats] = useState<RevenueStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const [formData, setFormData] = useState<InvoiceFormData>({
    user_id: 0,
    description: '',
    subtotal: 0,
    tax: 0,
    discount: 0,
    notes: '',
  });

  const [paymentData, setPaymentData] = useState<PaymentFormData>({
    invoice_id: 0,
    payment_gateway: 'stripe',
    amount: 0,
    payment_details: {},
  });

  // Mock data for demonstration
  const mockInvoices: Invoice[] = [
    {
      id: 1,
      invoice_number: 'INV-20250106-abc123de',
      user_id: 1,
      username: 'john_doe',
      email: 'john@example.com',
      package_id: 1,
      package_name: 'Premium IPTV',
      description: 'Monthly subscription - Premium IPTV',
      subtotal: 29.99,
      tax: 2.70,
      discount: 0,
      total: 32.69,
      status: 'paid',
      issue_date: '2025-01-06',
      due_date: '2025-01-13',
      paid_at: '2025-01-06T10:30:00Z',
      payment_method: 'gateway',
      payment_gateway: 'stripe',
      transaction_id: 'txn_abc123',
      created_at: '2025-01-06T08:00:00Z',
    },
    {
      id: 2,
      invoice_number: 'INV-20250105-xyz789fg',
      user_id: 2,
      username: 'jane_smith',
      email: 'jane@example.com',
      package_id: 2,
      package_name: 'Standard IPTV',
      description: 'Monthly subscription - Standard IPTV',
      subtotal: 19.99,
      tax: 1.80,
      discount: 2.00,
      total: 19.79,
      status: 'pending',
      issue_date: '2025-01-05',
      due_date: '2025-01-12',
      created_at: '2025-01-05T09:00:00Z',
    },
    {
      id: 3,
      invoice_number: 'INV-20250104-mno456pq',
      user_id: 3,
      username: 'mike_wilson',
      email: 'mike@example.com',
      package_id: 3,
      package_name: 'Enterprise IPTV',
      description: 'Monthly subscription - Enterprise IPTV',
      subtotal: 99.99,
      tax: 9.00,
      discount: 10.00,
      total: 98.99,
      status: 'overdue',
      issue_date: '2024-12-20',
      due_date: '2024-12-27',
      created_at: '2024-12-20T10:00:00Z',
    },
    {
      id: 4,
      invoice_number: 'INV-20250103-rst123uv',
      user_id: 4,
      username: 'sarah_jones',
      email: 'sarah@example.com',
      package_id: 1,
      package_name: 'Premium IPTV',
      description: 'Monthly subscription - Premium IPTV',
      subtotal: 29.99,
      tax: 2.70,
      discount: 0,
      total: 32.69,
      status: 'cancelled',
      issue_date: '2025-01-03',
      created_at: '2025-01-03T11:00:00Z',
    },
  ];

  const mockStats: RevenueStats = {
    total_revenue: 87654.32,
    monthly_revenue: 12450.00,
    weekly_revenue: 3280.50,
    daily_revenue: 875.25,
    pending_amount: 1250.00,
    overdue_amount: 485.75,
    total_invoices: 1247,
    paid_invoices: 1089,
    pending_invoices: 132,
    overdue_invoices: 26,
  };

  useEffect(() => {
    loadInvoices();
    loadStats();
  }, [page, searchTerm, statusFilter]);

  const loadInvoices = async () => {
    try {
      setLoading(true);
      // In production: const response = await billingAPI.listInvoices({ page, search: searchTerm, status: statusFilter });
      // For now, use mock data
      setTimeout(() => {
        let filtered = mockInvoices;
        if (searchTerm) {
          filtered = filtered.filter(inv =>
            inv.invoice_number.toLowerCase().includes(searchTerm.toLowerCase()) ||
            inv.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
            inv.email.toLowerCase().includes(searchTerm.toLowerCase())
          );
        }
        if (statusFilter) {
          filtered = filtered.filter(inv => inv.status === statusFilter);
        }
        setInvoices(filtered);
        setTotalPages(1);
        setLoading(false);
      }, 800);
    } catch (error) {
      console.error('Failed to load invoices:', error);
      setLoading(false);
    }
  };

  const loadStats = async () => {
    try {
      // In production: const response = await billingAPI.getRevenueStats();
      setStats(mockStats);
    } catch (error) {
      console.error('Failed to load stats:', error);
    }
  };

  const handleCreateInvoice = async () => {
    try {
      // Validate form
      if (!formData.user_id || formData.subtotal <= 0) {
        alert('Please fill in all required fields');
        return;
      }

      // In production: await billingAPI.createInvoice(formData);
      console.log('Creating invoice:', formData);

      alert('Invoice created successfully!');
      setShowCreateModal(false);
      resetForm();
      loadInvoices();
      loadStats();
    } catch (error) {
      console.error('Failed to create invoice:', error);
      alert('Failed to create invoice');
    }
  };

  const handleProcessPayment = async () => {
    try {
      if (!paymentData.invoice_id || paymentData.amount <= 0) {
        alert('Please fill in all required fields');
        return;
      }

      // In production: await billingAPI.processPayment(paymentData);
      console.log('Processing payment:', paymentData);

      alert('Payment processed successfully!');
      setShowPaymentModal(false);
      setSelectedInvoice(null);
      loadInvoices();
      loadStats();
    } catch (error) {
      console.error('Failed to process payment:', error);
      alert('Failed to process payment');
    }
  };

  const handleCancelInvoice = async (id: number) => {
    if (!confirm('Are you sure you want to cancel this invoice?')) return;

    try {
      // In production: await billingAPI.cancelInvoice(id);
      console.log('Cancelling invoice:', id);
      alert('Invoice cancelled successfully!');
      loadInvoices();
      loadStats();
    } catch (error) {
      console.error('Failed to cancel invoice:', error);
      alert('Failed to cancel invoice');
    }
  };

  const handleDeleteInvoice = async (id: number) => {
    if (!confirm('Are you sure you want to delete this invoice? This action cannot be undone.')) return;

    try {
      // In production: await billingAPI.deleteInvoice(id);
      console.log('Deleting invoice:', id);
      alert('Invoice deleted successfully!');
      loadInvoices();
      loadStats();
    } catch (error) {
      console.error('Failed to delete invoice:', error);
      alert('Failed to delete invoice');
    }
  };

  const openPaymentModal = (invoice: Invoice) => {
    setSelectedInvoice(invoice);
    setPaymentData({
      invoice_id: invoice.id,
      payment_gateway: 'stripe',
      amount: invoice.total,
      payment_details: {},
    });
    setShowPaymentModal(true);
  };

  const openViewModal = (invoice: Invoice) => {
    setSelectedInvoice(invoice);
    setShowViewModal(true);
  };

  const resetForm = () => {
    setFormData({
      user_id: 0,
      description: '',
      subtotal: 0,
      tax: 0,
      discount: 0,
      notes: '',
    });
  };

  const getStatusBadge = (status: string) => {
    const badges = {
      paid: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-700 dark:text-green-400', icon: CheckCircleIcon },
      pending: { bg: 'bg-yellow-100 dark:bg-yellow-900/30', text: 'text-yellow-700 dark:text-yellow-400', icon: ClockIcon },
      overdue: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-700 dark:text-red-400', icon: XCircleIcon },
      cancelled: { bg: 'bg-gray-100 dark:bg-gray-900/30', text: 'text-gray-700 dark:text-gray-400', icon: XCircleIcon },
      refunded: { bg: 'bg-blue-100 dark:bg-blue-900/30', text: 'text-blue-700 dark:text-blue-400', icon: ArrowPathIcon },
    };

    const badge = badges[status as keyof typeof badges] || badges.pending;
    const Icon = badge.icon;

    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${badge.bg} ${badge.text}`}>
        <Icon className="w-3 h-3 mr-1" />
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
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
              <span className="text-sm text-gray-500">vs last month</span>
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
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <StatCard
            title="Total Revenue"
            value={`$${stats.total_revenue.toLocaleString()}`}
            icon={BanknotesIcon}
            color="bg-gradient-to-br from-green-500 to-green-700"
          />
          <StatCard
            title="Monthly Revenue"
            value={`$${stats.monthly_revenue.toLocaleString()}`}
            icon={CreditCardIcon}
            color="bg-gradient-to-br from-blue-500 to-blue-700"
          />
          <StatCard
            title="Pending Amount"
            value={`$${stats.pending_amount.toLocaleString()}`}
            icon={ClockIcon}
            color="bg-gradient-to-br from-yellow-500 to-yellow-700"
          />
          <StatCard
            title="Paid Invoices"
            value={stats.paid_invoices}
            icon={CheckCircleIcon}
            color="bg-gradient-to-br from-purple-500 to-purple-700"
          />
        </div>
      )}

      {/* Filters and Actions */}
      <div className="card p-6">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
          <div className="flex-1 flex flex-col sm:flex-row gap-3">
            {/* Search */}
            <div className="relative flex-1">
              <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search by invoice number, username, or email..."
                className="input pl-10 w-full"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            {/* Status Filter */}
            <div className="relative sm:w-48">
              <FunnelIcon className="h-5 w-5 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
              <select
                className="input pl-10 w-full"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
              >
                <option value="">All Status</option>
                <option value="pending">Pending</option>
                <option value="paid">Paid</option>
                <option value="overdue">Overdue</option>
                <option value="cancelled">Cancelled</option>
                <option value="refunded">Refunded</option>
              </select>
            </div>
          </div>

          <button
            onClick={() => setShowCreateModal(true)}
            className="btn-primary"
          >
            <PlusIcon className="h-5 w-5 mr-2" />
            Create Invoice
          </button>
        </div>
      </div>

      {/* Invoices Table */}
      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th>Invoice Number</th>
                <th>Customer</th>
                <th>Package</th>
                <th>Amount</th>
                <th>Status</th>
                <th>Issue Date</th>
                <th>Due Date</th>
                <th>Actions</th>
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
              ) : invoices.length === 0 ? (
                <tr>
                  <td colSpan={8} className="text-center py-12 text-gray-500">
                    <DocumentTextIcon className="h-12 w-12 mx-auto mb-3 opacity-30" />
                    <p>No invoices found</p>
                  </td>
                </tr>
              ) : (
                invoices.map((invoice) => (
                  <tr key={invoice.id}>
                    <td className="font-mono text-sm">{invoice.invoice_number}</td>
                    <td>
                      <div>
                        <p className="font-medium text-gray-900 dark:text-white">{invoice.username}</p>
                        <p className="text-sm text-gray-500">{invoice.email}</p>
                      </div>
                    </td>
                    <td>{invoice.package_name || '-'}</td>
                    <td className="font-semibold">${invoice.total.toFixed(2)}</td>
                    <td>{getStatusBadge(invoice.status)}</td>
                    <td>{new Date(invoice.issue_date).toLocaleDateString()}</td>
                    <td>{invoice.due_date ? new Date(invoice.due_date).toLocaleDateString() : '-'}</td>
                    <td>
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={() => openViewModal(invoice)}
                          className="btn-icon"
                          title="View Details"
                        >
                          <EyeIcon className="h-4 w-4" />
                        </button>
                        {invoice.status === 'pending' && (
                          <>
                            <button
                              onClick={() => openPaymentModal(invoice)}
                              className="btn-icon text-green-600 hover:bg-green-50 dark:hover:bg-green-900/30"
                              title="Process Payment"
                            >
                              <CreditCardIcon className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => handleCancelInvoice(invoice.id)}
                              className="btn-icon text-yellow-600 hover:bg-yellow-50 dark:hover:bg-yellow-900/30"
                              title="Cancel Invoice"
                            >
                              <XCircleIcon className="h-4 w-4" />
                            </button>
                          </>
                        )}
                        {(invoice.status === 'pending' || invoice.status === 'cancelled') && (
                          <button
                            onClick={() => handleDeleteInvoice(invoice.id)}
                            className="btn-icon text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30"
                            title="Delete Invoice"
                          >
                            <TrashIcon className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {!loading && totalPages > 1 && (
          <div className="px-6 py-4 border-t border-gray-200 dark:border-dark-700 flex items-center justify-between">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="btn-secondary disabled:opacity-50"
            >
              Previous
            </button>
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Page {page} of {totalPages}
            </span>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="btn-secondary disabled:opacity-50"
            >
              Next
            </button>
          </div>
        )}
      </div>

      {/* Create Invoice Modal */}
      {showCreateModal && (
        <div className="modal-overlay">
          <div className="modal-content max-w-2xl">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Create New Invoice</h2>
              <button onClick={() => setShowCreateModal(false)} className="btn-icon">
                <XCircleIcon className="h-6 w-6" />
              </button>
            </div>

            <form className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="label">User ID *</label>
                  <input
                    type="number"
                    className="input"
                    value={formData.user_id || ''}
                    onChange={(e) => setFormData({ ...formData, user_id: parseInt(e.target.value) })}
                    required
                    min="1"
                  />
                </div>

                <div>
                  <label className="label">Package ID</label>
                  <input
                    type="number"
                    className="input"
                    value={formData.package_id || ''}
                    onChange={(e) => setFormData({ ...formData, package_id: parseInt(e.target.value) || undefined })}
                    min="1"
                  />
                </div>
              </div>

              <div>
                <label className="label">Description *</label>
                <input
                  type="text"
                  className="input"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  required
                  placeholder="Monthly subscription - Premium IPTV"
                />
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="label">Subtotal ($) *</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="input"
                    value={formData.subtotal || ''}
                    onChange={(e) => setFormData({ ...formData, subtotal: parseFloat(e.target.value) })}
                    required
                  />
                </div>

                <div>
                  <label className="label">Tax ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="input"
                    value={formData.tax || ''}
                    onChange={(e) => setFormData({ ...formData, tax: parseFloat(e.target.value) || 0 })}
                  />
                </div>

                <div>
                  <label className="label">Discount ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="input"
                    value={formData.discount || ''}
                    onChange={(e) => setFormData({ ...formData, discount: parseFloat(e.target.value) || 0 })}
                  />
                </div>
              </div>

              <div>
                <label className="label">Total</label>
                <div className="text-2xl font-bold text-gray-900 dark:text-white">
                  ${(formData.subtotal + formData.tax - formData.discount).toFixed(2)}
                </div>
              </div>

              <div>
                <label className="label">Due Date</label>
                <input
                  type="date"
                  className="input"
                  value={formData.due_date || ''}
                  onChange={(e) => setFormData({ ...formData, due_date: e.target.value || undefined })}
                />
              </div>

              <div>
                <label className="label">Notes</label>
                <textarea
                  className="input"
                  rows={3}
                  value={formData.notes}
                  onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  placeholder="Additional notes or instructions..."
                />
              </div>

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleCreateInvoice}
                  className="btn-primary"
                >
                  Create Invoice
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Payment Processing Modal */}
      {showPaymentModal && selectedInvoice && (
        <div className="modal-overlay">
          <div className="modal-content max-w-lg">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Process Payment</h2>
              <button onClick={() => setShowPaymentModal(false)} className="btn-icon">
                <XCircleIcon className="h-6 w-6" />
              </button>
            </div>

            <div className="mb-6 p-4 bg-gray-50 dark:bg-dark-800 rounded-lg">
              <div className="flex justify-between mb-2">
                <span className="text-gray-600 dark:text-gray-400">Invoice:</span>
                <span className="font-mono font-semibold">{selectedInvoice.invoice_number}</span>
              </div>
              <div className="flex justify-between mb-2">
                <span className="text-gray-600 dark:text-gray-400">Customer:</span>
                <span className="font-semibold">{selectedInvoice.username}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Amount Due:</span>
                <span className="text-2xl font-bold text-green-600 dark:text-green-400">
                  ${selectedInvoice.total.toFixed(2)}
                </span>
              </div>
            </div>

            <form className="space-y-4">
              <div>
                <label className="label">Payment Gateway *</label>
                <select
                  className="input"
                  value={paymentData.payment_gateway}
                  onChange={(e) => setPaymentData({ ...paymentData, payment_gateway: e.target.value as any })}
                  required
                >
                  <option value="stripe">Stripe</option>
                  <option value="paypal">PayPal</option>
                  <option value="crypto">Cryptocurrency</option>
                  <option value="bank_transfer">Bank Transfer</option>
                </select>
              </div>

              <div>
                <label className="label">Amount ($) *</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  className="input"
                  value={paymentData.amount || ''}
                  onChange={(e) => setPaymentData({ ...paymentData, amount: parseFloat(e.target.value) })}
                  required
                />
              </div>

              {paymentData.payment_gateway === 'stripe' && (
                <div className="space-y-3">
                  <div>
                    <label className="label">Card Number</label>
                    <input type="text" className="input" placeholder="4242 4242 4242 4242" />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="label">Expiry Date</label>
                      <input type="text" className="input" placeholder="MM/YY" />
                    </div>
                    <div>
                      <label className="label">CVV</label>
                      <input type="text" className="input" placeholder="123" />
                    </div>
                  </div>
                </div>
              )}

              {paymentData.payment_gateway === 'paypal' && (
                <div>
                  <label className="label">PayPal Email</label>
                  <input type="email" className="input" placeholder="customer@paypal.com" />
                </div>
              )}

              {paymentData.payment_gateway === 'crypto' && (
                <div>
                  <label className="label">Wallet Address</label>
                  <input type="text" className="input" placeholder="0x..." />
                </div>
              )}

              {paymentData.payment_gateway === 'bank_transfer' && (
                <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg border border-yellow-200 dark:border-yellow-800">
                  <p className="text-sm text-yellow-800 dark:text-yellow-200">
                    Payment will be marked as pending until bank transfer is verified manually.
                  </p>
                </div>
              )}

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowPaymentModal(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleProcessPayment}
                  className="btn-primary"
                >
                  <CreditCardIcon className="h-5 w-5 mr-2" />
                  Process Payment
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* View Invoice Modal */}
      {showViewModal && selectedInvoice && (
        <div className="modal-overlay">
          <div className="modal-content max-w-3xl">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Invoice Details</h2>
              <button onClick={() => setShowViewModal(false)} className="btn-icon">
                <XCircleIcon className="h-6 w-6" />
              </button>
            </div>

            <div className="space-y-6">
              {/* Invoice Header */}
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
                    {selectedInvoice.invoice_number}
                  </h3>
                  <p className="text-gray-600 dark:text-gray-400">
                    Issue Date: {new Date(selectedInvoice.issue_date).toLocaleDateString()}
                  </p>
                  {selectedInvoice.due_date && (
                    <p className="text-gray-600 dark:text-gray-400">
                      Due Date: {new Date(selectedInvoice.due_date).toLocaleDateString()}
                    </p>
                  )}
                </div>
                <div>{getStatusBadge(selectedInvoice.status)}</div>
              </div>

              {/* Customer Info */}
              <div className="grid grid-cols-2 gap-6">
                <div>
                  <h4 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-2">BILL TO</h4>
                  <p className="font-semibold text-gray-900 dark:text-white">{selectedInvoice.username}</p>
                  <p className="text-gray-600 dark:text-gray-400">{selectedInvoice.email}</p>
                  <p className="text-sm text-gray-500">User ID: {selectedInvoice.user_id}</p>
                </div>
                {selectedInvoice.package_name && (
                  <div>
                    <h4 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-2">PACKAGE</h4>
                    <p className="font-semibold text-gray-900 dark:text-white">{selectedInvoice.package_name}</p>
                  </div>
                )}
              </div>

              {/* Line Items */}
              <div>
                <h4 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-3">DESCRIPTION</h4>
                <p className="text-gray-900 dark:text-white">{selectedInvoice.description}</p>
              </div>

              {/* Amounts */}
              <div className="border-t border-gray-200 dark:border-dark-700 pt-4">
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-gray-600 dark:text-gray-400">Subtotal:</span>
                    <span className="font-semibold">${selectedInvoice.subtotal.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600 dark:text-gray-400">Tax:</span>
                    <span className="font-semibold">${selectedInvoice.tax.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600 dark:text-gray-400">Discount:</span>
                    <span className="font-semibold text-green-600">-${selectedInvoice.discount.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between text-xl border-t border-gray-200 dark:border-dark-700 pt-2 mt-2">
                    <span className="font-bold text-gray-900 dark:text-white">Total:</span>
                    <span className="font-bold text-gray-900 dark:text-white">${selectedInvoice.total.toFixed(2)}</span>
                  </div>
                </div>
              </div>

              {/* Payment Info */}
              {selectedInvoice.status === 'paid' && (
                <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg border border-green-200 dark:border-green-800">
                  <h4 className="text-sm font-semibold text-green-800 dark:text-green-300 mb-2">PAYMENT INFORMATION</h4>
                  <div className="space-y-1 text-sm">
                    <p className="text-green-700 dark:text-green-400">
                      Paid on: {selectedInvoice.paid_at ? new Date(selectedInvoice.paid_at).toLocaleString() : '-'}
                    </p>
                    <p className="text-green-700 dark:text-green-400">
                      Method: {selectedInvoice.payment_gateway?.toUpperCase()}
                    </p>
                    {selectedInvoice.transaction_id && (
                      <p className="text-green-700 dark:text-green-400 font-mono">
                        Transaction ID: {selectedInvoice.transaction_id}
                      </p>
                    )}
                  </div>
                </div>
              )}

              {/* Notes */}
              {selectedInvoice.notes && (
                <div>
                  <h4 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-2">NOTES</h4>
                  <p className="text-gray-700 dark:text-gray-300">{selectedInvoice.notes}</p>
                </div>
              )}

              <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200 dark:border-dark-700">
                <button onClick={() => setShowViewModal(false)} className="btn-secondary">
                  Close
                </button>
                {selectedInvoice.status === 'pending' && (
                  <button
                    onClick={() => {
                      setShowViewModal(false);
                      openPaymentModal(selectedInvoice);
                    }}
                    className="btn-primary"
                  >
                    <CreditCardIcon className="h-5 w-5 mr-2" />
                    Process Payment
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
