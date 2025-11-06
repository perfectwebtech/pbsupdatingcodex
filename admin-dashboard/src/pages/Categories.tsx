import React, { useState, useEffect } from 'react';
import {
  FolderIcon,
  FolderOpenIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  TrashIcon,
  PencilIcon,
  ChevronRightIcon,
  ChevronDownIcon,
} from '@heroicons/react/24/outline';

interface Category {
  id: number;
  name: string;
  parent_id: number | null;
  description?: string;
  icon?: string;
  stream_count: number;
  display_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface CategoryStats {
  total_categories: number;
  parent_categories: number;
  child_categories: number;
  total_streams: number;
}

interface CategoryFormData {
  name: string;
  parent_id: number | null;
  description?: string;
  icon?: string;
  display_order: number;
  is_active: boolean;
}

interface CategoryNode extends Category {
  children: CategoryNode[];
  isExpanded?: boolean;
}

const Categories: React.FC = () => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [categoryTree, setCategoryTree] = useState<CategoryNode[]>([]);
  const [expandedCategories, setExpandedCategories] = useState<Set<number>>(new Set());
  const [stats, setStats] = useState<CategoryStats>({
    total_categories: 0,
    parent_categories: 0,
    child_categories: 0,
    total_streams: 0,
  });

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive'>('all');

  const [formData, setFormData] = useState<CategoryFormData>({
    name: '',
    parent_id: null,
    description: '',
    icon: '',
    display_order: 0,
    is_active: true,
  });

  // Mock data for demonstration
  useEffect(() => {
    loadMockData();
  }, []);

  useEffect(() => {
    // Build tree whenever categories change
    const tree = buildCategoryTree(categories);
    setCategoryTree(tree);
  }, [categories, expandedCategories]);

  const loadMockData = () => {
    const mockCategories: Category[] = [
      // Parent Categories
      {
        id: 1,
        name: 'News',
        parent_id: null,
        description: 'News channels from around the world',
        icon: '📰',
        stream_count: 25,
        display_order: 1,
        is_active: true,
        created_at: new Date(Date.now() - 180 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 1 * 86400000).toISOString(),
      },
      {
        id: 2,
        name: 'Sports',
        parent_id: null,
        description: 'Sports channels and events',
        icon: '⚽',
        stream_count: 45,
        display_order: 2,
        is_active: true,
        created_at: new Date(Date.now() - 175 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 2 * 86400000).toISOString(),
      },
      {
        id: 3,
        name: 'Entertainment',
        parent_id: null,
        description: 'Movies, TV shows, and entertainment',
        icon: '🎬',
        stream_count: 120,
        display_order: 3,
        is_active: true,
        created_at: new Date(Date.now() - 170 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 3 * 86400000).toISOString(),
      },
      {
        id: 4,
        name: 'Documentary',
        parent_id: null,
        description: 'Educational and documentary content',
        icon: '🎓',
        stream_count: 35,
        display_order: 4,
        is_active: true,
        created_at: new Date(Date.now() - 165 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 4 * 86400000).toISOString(),
      },
      {
        id: 5,
        name: 'Kids',
        parent_id: null,
        description: 'Family-friendly content for children',
        icon: '👶',
        stream_count: 28,
        display_order: 5,
        is_active: true,
        created_at: new Date(Date.now() - 160 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 5 * 86400000).toISOString(),
      },

      // Child Categories - News
      {
        id: 11,
        name: 'International News',
        parent_id: 1,
        description: 'Global news coverage',
        icon: '🌍',
        stream_count: 12,
        display_order: 1,
        is_active: true,
        created_at: new Date(Date.now() - 150 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 6 * 86400000).toISOString(),
      },
      {
        id: 12,
        name: 'Business News',
        parent_id: 1,
        description: 'Financial and business news',
        icon: '💼',
        stream_count: 8,
        display_order: 2,
        is_active: true,
        created_at: new Date(Date.now() - 145 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 7 * 86400000).toISOString(),
      },
      {
        id: 13,
        name: 'Technology News',
        parent_id: 1,
        description: 'Tech industry news and updates',
        icon: '💻',
        stream_count: 5,
        display_order: 3,
        is_active: true,
        created_at: new Date(Date.now() - 140 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 8 * 86400000).toISOString(),
      },

      // Child Categories - Sports
      {
        id: 21,
        name: 'Football',
        parent_id: 2,
        description: 'Soccer and football leagues',
        icon: '⚽',
        stream_count: 18,
        display_order: 1,
        is_active: true,
        created_at: new Date(Date.now() - 135 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 9 * 86400000).toISOString(),
      },
      {
        id: 22,
        name: 'Basketball',
        parent_id: 2,
        description: 'NBA, NCAA, and international basketball',
        icon: '🏀',
        stream_count: 12,
        display_order: 2,
        is_active: true,
        created_at: new Date(Date.now() - 130 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 10 * 86400000).toISOString(),
      },
      {
        id: 23,
        name: 'Racing',
        parent_id: 2,
        description: 'F1, NASCAR, and motorsports',
        icon: '🏎️',
        stream_count: 8,
        display_order: 3,
        is_active: true,
        created_at: new Date(Date.now() - 125 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 11 * 86400000).toISOString(),
      },
      {
        id: 24,
        name: 'Combat Sports',
        parent_id: 2,
        description: 'MMA, Boxing, Wrestling',
        icon: '🥊',
        stream_count: 7,
        display_order: 4,
        is_active: true,
        created_at: new Date(Date.now() - 120 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 12 * 86400000).toISOString(),
      },

      // Child Categories - Entertainment
      {
        id: 31,
        name: 'Action Movies',
        parent_id: 3,
        description: 'Action and adventure films',
        icon: '💥',
        stream_count: 35,
        display_order: 1,
        is_active: true,
        created_at: new Date(Date.now() - 115 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 13 * 86400000).toISOString(),
      },
      {
        id: 32,
        name: 'Comedy',
        parent_id: 3,
        description: 'Comedy shows and movies',
        icon: '😂',
        stream_count: 28,
        display_order: 2,
        is_active: true,
        created_at: new Date(Date.now() - 110 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 14 * 86400000).toISOString(),
      },
      {
        id: 33,
        name: 'Drama',
        parent_id: 3,
        description: 'Drama series and films',
        icon: '🎭',
        stream_count: 32,
        display_order: 3,
        is_active: true,
        created_at: new Date(Date.now() - 105 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 15 * 86400000).toISOString(),
      },
      {
        id: 34,
        name: 'Horror',
        parent_id: 3,
        description: 'Horror and thriller content',
        icon: '👻',
        stream_count: 25,
        display_order: 4,
        is_active: true,
        created_at: new Date(Date.now() - 100 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 16 * 86400000).toISOString(),
      },

      // Child Categories - Documentary
      {
        id: 41,
        name: 'Nature & Wildlife',
        parent_id: 4,
        description: 'Nature documentaries',
        icon: '🦁',
        stream_count: 15,
        display_order: 1,
        is_active: true,
        created_at: new Date(Date.now() - 95 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 17 * 86400000).toISOString(),
      },
      {
        id: 42,
        name: 'Science & Tech',
        parent_id: 4,
        description: 'Science and technology documentaries',
        icon: '🔬',
        stream_count: 12,
        display_order: 2,
        is_active: true,
        created_at: new Date(Date.now() - 90 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 18 * 86400000).toISOString(),
      },
      {
        id: 43,
        name: 'History',
        parent_id: 4,
        description: 'Historical documentaries',
        icon: '📜',
        stream_count: 8,
        display_order: 3,
        is_active: true,
        created_at: new Date(Date.now() - 85 * 86400000).toISOString(),
        updated_at: new Date(Date.now() - 19 * 86400000).toISOString(),
      },
    ];

    const mockStats: CategoryStats = {
      total_categories: mockCategories.length,
      parent_categories: mockCategories.filter(c => c.parent_id === null).length,
      child_categories: mockCategories.filter(c => c.parent_id !== null).length,
      total_streams: mockCategories.reduce((sum, c) => sum + c.stream_count, 0),
    };

    setCategories(mockCategories);
    setStats(mockStats);

    // Auto-expand all parent categories by default
    const parentIds = new Set(mockCategories.filter(c => c.parent_id === null).map(c => c.id));
    setExpandedCategories(parentIds);
  };

  const buildCategoryTree = (cats: Category[]): CategoryNode[] => {
    const categoryMap = new Map<number, CategoryNode>();
    const rootCategories: CategoryNode[] = [];

    // Create nodes
    cats.forEach(cat => {
      categoryMap.set(cat.id, {
        ...cat,
        children: [],
        isExpanded: expandedCategories.has(cat.id),
      });
    });

    // Build tree
    cats.forEach(cat => {
      const node = categoryMap.get(cat.id)!;
      if (cat.parent_id === null) {
        rootCategories.push(node);
      } else {
        const parent = categoryMap.get(cat.parent_id);
        if (parent) {
          parent.children.push(node);
        }
      }
    });

    // Sort by display_order
    const sortByOrder = (a: CategoryNode, b: CategoryNode) => a.display_order - b.display_order;
    rootCategories.sort(sortByOrder);
    rootCategories.forEach(root => {
      root.children.sort(sortByOrder);
    });

    return rootCategories;
  };

  const toggleCategory = (id: number) => {
    const newExpanded = new Set(expandedCategories);
    if (newExpanded.has(id)) {
      newExpanded.delete(id);
    } else {
      newExpanded.add(id);
    }
    setExpandedCategories(newExpanded);
  };

  const handleCreateCategory = async () => {
    if (!formData.name) {
      alert('Please fill in all required fields');
      return;
    }

    // In production: await categoryAPI.createCategory(formData);
    console.log('Creating category:', formData);

    // Mock: Add to list
    const newCategory: Category = {
      id: categories.length + 1,
      ...formData,
      stream_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setCategories([...categories, newCategory]);

    setShowCreateModal(false);
    resetForm();
    alert('Category created successfully!');
  };

  const handleEditCategory = async () => {
    if (!selectedCategory || !formData.name) {
      alert('Please fill in all required fields');
      return;
    }

    // In production: await categoryAPI.updateCategory(selectedCategory.id, formData);
    console.log('Updating category:', selectedCategory.id, formData);

    // Mock: Update category
    setCategories(categories.map(c =>
      c.id === selectedCategory.id
        ? { ...c, ...formData, updated_at: new Date().toISOString() }
        : c
    ));

    setShowEditModal(false);
    setSelectedCategory(null);
    resetForm();
    alert('Category updated successfully!');
  };

  const handleDeleteCategory = async (category: Category) => {
    // Check if has children
    const hasChildren = categories.some(c => c.parent_id === category.id);
    if (hasChildren) {
      alert('Cannot delete category with subcategories. Please delete or reassign subcategories first.');
      return;
    }

    if (!confirm(`Delete category "${category.name}"?`)) return;

    // In production: await categoryAPI.deleteCategory(category.id);
    console.log('Deleting category:', category.id);

    // Mock: Remove category
    setCategories(categories.filter(c => c.id !== category.id));
    alert('Category deleted successfully!');
  };

  const resetForm = () => {
    setFormData({
      name: '',
      parent_id: null,
      description: '',
      icon: '',
      display_order: 0,
      is_active: true,
    });
  };

  const openEditModal = (category: Category) => {
    setSelectedCategory(category);
    setFormData({
      name: category.name,
      parent_id: category.parent_id,
      description: category.description,
      icon: category.icon,
      display_order: category.display_order,
      is_active: category.is_active,
    });
    setShowEditModal(true);
  };

  const renderCategory = (category: CategoryNode, level: number = 0) => {
    const hasChildren = category.children.length > 0;
    const isExpanded = category.isExpanded;

    // Apply filters
    const matchesSearch = searchTerm === '' ||
      category.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      category.description?.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesStatus = filterStatus === 'all' ||
      (filterStatus === 'active' && category.is_active) ||
      (filterStatus === 'inactive' && !category.is_active);

    if (!matchesSearch || !matchesStatus) return null;

    const paddingLeft = level * 32 + 16;

    return (
      <React.Fragment key={category.id}>
        {/* Category Row */}
        <tr className="hover:bg-gray-50 dark:hover:bg-gray-700">
          <td className="px-6 py-4 whitespace-nowrap" style={{ paddingLeft: `${paddingLeft}px` }}>
            <div className="flex items-center gap-2">
              {hasChildren && (
                <button
                  onClick={() => toggleCategory(category.id)}
                  className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                >
                  {isExpanded ? (
                    <ChevronDownIcon className="w-4 h-4" />
                  ) : (
                    <ChevronRightIcon className="w-4 h-4" />
                  )}
                </button>
              )}
              {!hasChildren && <div className="w-4" />}
              <span className="text-2xl">{category.icon || (hasChildren ? '📁' : '📄')}</span>
              <div>
                <div className="text-sm font-medium text-gray-900 dark:text-white">
                  {category.name}
                </div>
                {category.description && (
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    {category.description}
                  </div>
                )}
              </div>
            </div>
          </td>
          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
            {category.stream_count} streams
          </td>
          <td className="px-6 py-4 whitespace-nowrap">
            {category.is_active ? (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                Active
              </span>
            ) : (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300">
                Inactive
              </span>
            )}
          </td>
          <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
            <div className="flex items-center justify-end gap-2">
              <button
                onClick={() => openEditModal(category)}
                className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
                title="Edit"
              >
                <PencilIcon className="w-5 h-5" />
              </button>
              <button
                onClick={() => handleDeleteCategory(category)}
                className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                title="Delete"
              >
                <TrashIcon className="w-5 h-5" />
              </button>
            </div>
          </td>
        </tr>

        {/* Children */}
        {isExpanded && hasChildren && category.children.map(child => renderCategory(child, level + 1))}
      </React.Fragment>
    );
  };

  return (
    <div className="p-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <FolderIcon className="w-8 h-8 text-blue-500" />
          Category Management
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          Organize content into hierarchical categories
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Total Categories</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-2">{stats.total_categories}</p>
            </div>
            <FolderIcon className="w-12 h-12 text-blue-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Parent Categories</p>
              <p className="text-3xl font-bold text-purple-600 dark:text-purple-400 mt-2">{stats.parent_categories}</p>
            </div>
            <FolderOpenIcon className="w-12 h-12 text-purple-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Subcategories</p>
              <p className="text-3xl font-bold text-green-600 dark:text-green-400 mt-2">{stats.child_categories}</p>
            </div>
            <FolderIcon className="w-12 h-12 text-green-500" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 dark:text-gray-400 text-sm">Total Streams</p>
              <p className="text-3xl font-bold text-orange-600 dark:text-orange-400 mt-2">{stats.total_streams}</p>
            </div>
            <div className="text-5xl">📺</div>
          </div>
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
                placeholder="Search categories..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              />
            </div>

            {/* Status Filter */}
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value as 'all' | 'active' | 'inactive')}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="all">All Status</option>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>

          <button
            onClick={() => setShowCreateModal(true)}
            className="w-full sm:w-auto bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg flex items-center gap-2 justify-center transition-colors"
          >
            <PlusIcon className="w-5 h-5" />
            Add Category
          </button>
        </div>
      </div>

      {/* Categories Tree Table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Category Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Streams
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {categoryTree.map(category => renderCategory(category, 0))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Category Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Create New Category</h2>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Category Name *
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="e.g., Sports, News, Movies"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Parent Category
                  </label>
                  <select
                    value={formData.parent_id || ''}
                    onChange={(e) => setFormData({ ...formData, parent_id: e.target.value ? parseInt(e.target.value) : null })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  >
                    <option value="">None (Top Level)</option>
                    {categories.filter(c => c.parent_id === null).map(cat => (
                      <option key={cat.id} value={cat.id}>{cat.name}</option>
                    ))}
                  </select>
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
                    placeholder="Category description..."
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Icon (Emoji)
                    </label>
                    <input
                      type="text"
                      value={formData.icon}
                      onChange={(e) => setFormData({ ...formData, icon: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="📺"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Display Order
                    </label>
                    <input
                      type="number"
                      value={formData.display_order}
                      onChange={(e) => setFormData({ ...formData, display_order: parseInt(e.target.value) })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="0"
                    />
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    Active
                  </label>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={handleCreateCategory}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Create Category
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

      {/* Edit Category Modal */}
      {showEditModal && selectedCategory && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">Edit Category</h2>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Category Name *
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="e.g., Sports, News, Movies"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Parent Category
                  </label>
                  <select
                    value={formData.parent_id || ''}
                    onChange={(e) => setFormData({ ...formData, parent_id: e.target.value ? parseInt(e.target.value) : null })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  >
                    <option value="">None (Top Level)</option>
                    {categories.filter(c => c.parent_id === null && c.id !== selectedCategory.id).map(cat => (
                      <option key={cat.id} value={cat.id}>{cat.name}</option>
                    ))}
                  </select>
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
                    placeholder="Category description..."
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Icon (Emoji)
                    </label>
                    <input
                      type="text"
                      value={formData.icon}
                      onChange={(e) => setFormData({ ...formData, icon: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="📺"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      Display Order
                    </label>
                    <input
                      type="number"
                      value={formData.display_order}
                      onChange={(e) => setFormData({ ...formData, display_order: parseInt(e.target.value) })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      placeholder="0"
                    />
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    Active
                  </label>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={handleEditCategory}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg transition-colors"
                >
                  Update Category
                </button>
                <button
                  onClick={() => {
                    setShowEditModal(false);
                    setSelectedCategory(null);
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

export default Categories;
