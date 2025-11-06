# 🧭 Menu Structure & Navigation

Complete documentation of the admin dashboard menu system and navigation.

**Last Updated:** January 6, 2025
**Version:** 1.0.0

---

## 📋 Admin Dashboard Menu

### 🏠 Main Navigation

```
┌─────────────────────────────────────────────────────────┐
│  IPTV Platform Admin Dashboard                         │
└─────────────────────────────────────────────────────────┘

📊 Dashboard                        /dashboard
   ├─ Overview statistics
   ├─ User growth chart
   ├─ Revenue chart
   ├─ Content distribution
   └─ Top 5 streams

👥 Users                           /users
   ├─ User list (table view)
   ├─ Create new user
   ├─ Edit user
   ├─ Manage subscriptions
   ├─ Activity logs
   └─ User statistics

📺 Streams                         /streams
   ├─ Live streams list
   ├─ VOD content list
   ├─ Create new stream
   ├─ Edit stream
   ├─ Stream analytics
   └─ Featured streams

📁 Categories                      /categories
   ├─ Category list
   ├─ Create category
   ├─ Edit category
   ├─ Category statistics
   └─ Category hierarchy

💳 Billing                         /billing
   ├─ Invoice list
   ├─ Create invoice
   ├─ Payment processing
   ├─ Transaction history
   ├─ Payment methods
   └─ Revenue statistics

👔 Resellers                       /resellers
   ├─ Reseller list
   ├─ Create reseller
   ├─ Edit reseller
   ├─ Credits management
   ├─ Customer assignments
   ├─ Commission tracking
   └─ Reseller statistics

🎬 Series                          /series
   ├─ Series list (grid view)
   ├─ Create series
   ├─ Edit series
   ├─ Episode management
   ├─ Season organization
   └─ Series analytics

📅 EPG                             /epg
   ├─ Program schedule
   ├─ EPG sources
   ├─ XMLTV import
   ├─ Program editor
   ├─ Import history
   └─ User reminders

📱 Devices                         /devices
   ├─ Device list
   ├─ Register device
   ├─ Device sessions
   ├─ Block/Unblock
   ├─ MAG portal
   └─ Enigma2 support

💎 Packages                        /packages
   ├─ Package list
   ├─ Create package
   ├─ Edit package
   ├─ Stream assignments
   └─ Package statistics

📊 Analytics                       /analytics
   ├─ Dashboard stats
   ├─ User analytics
   ├─ Stream analytics
   ├─ Revenue reports
   ├─ Popular content
   └─ Geographic data

🔄 Sessions                        /sessions
   ├─ Active sessions
   ├─ Session history
   ├─ Concurrent viewers
   ├─ Session analytics
   └─ Terminate sessions

🎨 Transcoding                     /transcoding
   ├─ Transcoding jobs
   ├─ Job queue
   ├─ Quality profiles
   ├─ Codec settings
   └─ Job statistics

⚙️ Settings                        /settings
   ├─ General settings
   ├─ Payment gateways
   ├─ Email templates
   ├─ System configuration
   ├─ Security settings
   └─ API keys
```

---

## 📱 Menu Implementation Status

### ✅ Fully Implemented (UI + Backend + Database)

| Menu Item | Route | Components | Backend | Database | Status |
|-----------|-------|------------|---------|----------|--------|
| **Dashboard** | `/dashboard` | ✅ Dashboard.tsx | ✅ Analytics API | ✅ All tables | ✅ Complete |
| **Users** | `/users` | ✅ Users.tsx | ✅ User handler | ✅ users table | ✅ Complete |
| **Billing** | `/billing` | ✅ Billing.tsx | ✅ billing_handler.go | ✅ invoices, transactions | ✅ Complete |
| **Resellers** | `/resellers` | ✅ Resellers.tsx | ✅ reseller_handler.go | ✅ resellers, credits_log | ✅ Complete |
| **Series** | `/series` | ✅ Series.tsx | ✅ series_handler.go | ✅ series, episodes | ✅ Complete |

### 🔄 Partially Implemented (Backend ready, UI pending)

| Menu Item | Route | Components | Backend | Database | Status |
|-----------|-------|------------|---------|----------|--------|
| **Streams** | `/streams` | 🔄 Placeholder | ✅ Stream handler | ✅ streams table | 🔄 Backend Ready |
| **Categories** | `/categories` | 🔄 Placeholder | ✅ Category handler | ✅ categories table | 🔄 Backend Ready |
| **Packages** | `/packages` | ❌ Not started | ✅ Package handler | ✅ packages table | 🔄 Backend Ready |

### 📝 Database Ready (Schema complete, needs handler + UI)

| Menu Item | Route | Components | Backend | Database | Status |
|-----------|-------|------------|---------|----------|--------|
| **EPG** | `/epg` | ❌ Not started | ❌ Not started | ✅ EPG tables | 📝 Database Ready |
| **Devices** | `/devices` | ❌ Not started | ❌ Not started | ✅ devices tables | 📝 Database Ready |

### ❌ Not Yet Implemented

| Menu Item | Route | Components | Backend | Database | Status |
|-----------|-------|------------|---------|----------|--------|
| **Analytics** | `/analytics` | ❌ Not started | 🔄 Partial | ✅ All tables | ❌ Needs implementation |
| **Sessions** | `/sessions` | ❌ Not started | ❌ Not started | 🔄 Partial | ❌ Needs implementation |
| **Transcoding** | `/transcoding` | ❌ Not started | ❌ Not started | ❌ Not started | ❌ Needs implementation |
| **Settings** | `/settings` | ❌ Not started | ❌ Not started | ❌ Not started | ❌ Needs implementation |

---

## 🎯 Menu Component Structure

### Current Implementation

```typescript
// admin-dashboard/src/components/Layout.tsx

const menuItems = [
  {
    name: 'Dashboard',
    href: '/dashboard',
    icon: ChartBarIcon,
    status: 'complete' // ✅
  },
  {
    name: 'Users',
    href: '/users',
    icon: UsersIcon,
    status: 'complete' // ✅
  },
  {
    name: 'Streams',
    href: '/streams',
    icon: PlayCircleIcon,
    status: 'partial' // 🔄
  },
  {
    name: 'Categories',
    href: '/categories',
    icon: FolderIcon,
    status: 'partial' // 🔄
  },
  {
    name: 'Billing',
    href: '/billing',
    icon: CreditCardIcon,
    status: 'complete' // ✅
  },
  {
    name: 'Resellers',
    href: '/resellers',
    icon: UserGroupIcon,
    status: 'complete' // ✅
  },
  {
    name: 'Series',
    href: '/series',
    icon: FilmIcon,
    status: 'complete' // ✅
  },
  {
    name: 'EPG',
    href: '/epg',
    icon: CalendarIcon,
    status: 'database-ready' // 📝
  },
  {
    name: 'Devices',
    href: '/devices',
    icon: DevicePhoneMobileIcon,
    status: 'database-ready' // 📝
  },
  {
    name: 'Analytics',
    href: '/analytics',
    icon: ChartPieIcon,
    status: 'not-started' // ❌
  },
  {
    name: 'Sessions',
    href: '/sessions',
    icon: SignalIcon,
    status: 'not-started' // ❌
  },
  {
    name: 'Transcoding',
    href: '/transcoding',
    icon: CogIcon,
    status: 'not-started' // ❌
  },
  {
    name: 'Settings',
    href: '/settings',
    icon: Cog6ToothIcon,
    status: 'not-started' // ❌
  }
];
```

---

## 🗂️ File Structure

```
admin-dashboard/src/
├── pages/
│   ├── Dashboard.tsx           ✅ Complete (800 lines)
│   ├── Users.tsx               ✅ Complete (600 lines)
│   ├── Billing.tsx             ✅ Complete (900 lines)
│   ├── Resellers.tsx           ✅ Complete (500 lines)
│   ├── Series.tsx              ✅ Complete (1000 lines)
│   ├── Streams.tsx             🔄 Placeholder (50 lines)
│   ├── Categories.tsx          🔄 Placeholder (50 lines)
│   ├── EPG.tsx                 ❌ Not created
│   ├── Devices.tsx             ❌ Not created
│   ├── Packages.tsx            ❌ Not created
│   ├── Analytics.tsx           🔄 Placeholder (50 lines)
│   ├── Sessions.tsx            🔄 Placeholder (50 lines)
│   ├── Transcoding.tsx         🔄 Placeholder (50 lines)
│   └── Settings.tsx            🔄 Placeholder (50 lines)
│
├── components/
│   ├── Layout.tsx              ✅ Complete (navigation)
│   ├── Sidebar.tsx             ✅ Complete (menu)
│   └── Header.tsx              ✅ Complete (user menu)
│
├── services/
│   └── api.ts                  ✅ Complete (axios setup)
│
└── stores/
    └── authStore.ts            ✅ Complete (Zustand)
```

---

## 🎨 Menu Features

### Current Features

#### ✅ Implemented
- [x] Responsive sidebar navigation
- [x] Active route highlighting
- [x] Icon support (Heroicons)
- [x] Dark mode support
- [x] Mobile menu (hamburger)
- [x] User dropdown menu
- [x] Logout functionality
- [x] Breadcrumb navigation
- [x] Page titles

#### 🔄 In Progress
- [ ] Menu item badges (counts)
- [ ] Nested menu items
- [ ] Menu search
- [ ] Favorite/pinned items
- [ ] Recent pages

#### ❌ Planned
- [ ] Role-based menu visibility
- [ ] Custom menu ordering
- [ ] Menu personalization
- [ ] Keyboard shortcuts
- [ ] Quick actions menu

---

## 🔐 Menu Access Control

### Role-Based Menu Items

```typescript
// Future implementation
const menuItemsByRole = {
  admin: [
    'dashboard', 'users', 'streams', 'categories',
    'billing', 'resellers', 'series', 'epg',
    'devices', 'analytics', 'sessions',
    'transcoding', 'settings'
  ],
  reseller: [
    'dashboard', 'users', 'billing', 'analytics'
  ],
  user: [
    'dashboard', 'streams', 'series', 'epg'
  ]
};
```

---

## 📊 Menu Usage Statistics (Planned)

Track which menu items are most used:

```sql
CREATE TABLE menu_analytics (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id),
  menu_item VARCHAR(50),
  accessed_at TIMESTAMP DEFAULT NOW()
);
```

---

## 🎯 Completion Progress

### Overall Progress

```
Total Menu Items: 13
✅ Fully Complete: 5 (38%)
🔄 Partial:        3 (23%)
📝 DB Ready:       2 (15%)
❌ Not Started:    3 (23%)
```

### Development Priority

**High Priority (Next):**
1. Streams management UI
2. Categories management UI
3. EPG handler + service + UI
4. Devices handler + service + UI

**Medium Priority:**
5. Enhanced Analytics dashboard
6. Sessions monitoring
7. Packages UI

**Low Priority:**
8. Transcoding UI
9. Settings page
10. Advanced features

---

## 📱 Mobile Menu

### Mobile Navigation Features

- [x] Hamburger menu icon
- [x] Slide-out drawer
- [x] Touch-friendly targets
- [x] Swipe gestures
- [x] Close on route change
- [x] Overlay backdrop

---

## 🎨 Menu Styling

### Current Theme

```css
/* Active menu item */
.menu-item-active {
  background: linear-gradient(to right, #3b82f6, #2563eb);
  color: white;
  font-weight: 600;
}

/* Hover state */
.menu-item:hover {
  background: rgba(59, 130, 246, 0.1);
  transform: translateX(4px);
  transition: all 0.2s ease;
}

/* Dark mode */
.dark .menu-item {
  color: #9ca3af;
}

.dark .menu-item-active {
  background: linear-gradient(to right, #1e40af, #1e3a8a);
  color: white;
}
```

---

## 🔔 Menu Notifications (Planned)

Future notification badges:

```typescript
const menuBadges = {
  users: { count: 3, type: 'new' },
  billing: { count: 12, type: 'pending' },
  sessions: { count: 145, type: 'active' }
};
```

---

## 🚀 Quick Actions Menu (Planned)

Global quick actions accessible from any page:

```
⌘K - Quick search
⌘N - New user
⌘I - New invoice
⌘S - New stream
⌘P - New package
```

---

**Note:** This menu structure is designed to match Xtream UI feature parity while providing a modern, intuitive user experience.
