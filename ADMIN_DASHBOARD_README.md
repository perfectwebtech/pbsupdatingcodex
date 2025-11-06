# 🎨 IPTV Platform - Enterprise Admin Dashboard

**Modern, Powerful, and Beautiful Admin Interface**

---

## 🌟 **OVERVIEW**

A comprehensive, production-ready admin dashboard built with **React 18**, **TypeScript**, **Tailwind CSS**, and **Vite**. Features a modern, responsive UI with dark mode support, real-time updates, and powerful data visualization.

---

## ✨ **KEY FEATURES**

### **Authentication & Security**
- ✅ JWT-based authentication with auto-refresh
- ✅ Secure token storage and management
- ✅ Protected routes and role-based access
- ✅ Session management
- ✅ Remember me functionality
- ✅ Password visibility toggle

### **User Interface**
- ✅ Modern, clean design with Tailwind CSS
- ✅ Dark mode support (auto-switching)
- ✅ Responsive design (mobile, tablet, desktop)
- ✅ Smooth animations and transitions
- ✅ Toast notifications (success/error/info)
- ✅ Loading states and skeletons
- ✅ Glass morphism effects
- ✅ Gradient accents and shadows

### **Dashboard Pages**
1. **Dashboard Home** - Real-time statistics and analytics
2. **Users Management** - CRUD operations for users
3. **Streams Management** - Manage live/VOD content
4. **Categories** - Organize content categories
5. **Analytics** - Charts and insights
6. **Live Sessions** - Monitor active viewers
7. **Transcoding** - Video processing jobs
8. **Settings** - System configuration

### **Components**
- **Layout** - Responsive sidebar and header
- **Stat Cards** - Animated statistics display
- **Data Tables** - Sortable, filterable tables
- **Charts** - Line, bar, pie charts (Recharts)
- **Modals** - Create/edit forms
- **Badges** - Status indicators
- **Buttons** - Primary, secondary, danger variants

### **API Integration**
- ✅ Axios-based HTTP client
- ✅ Request/response interceptors
- ✅ Automatic token refresh
- ✅ Error handling and retries
- ✅ API service layer organization

---

## 🚀 **TECHNOLOGY STACK**

### **Frontend**
```
- React 18.2          - Modern UI library
- TypeScript 5.3      - Type safety
- Vite 5.0            - Lightning-fast bundler
- Tailwind CSS 3.3    - Utility-first CSS
- React Router 6.20   - Client-side routing
```

### **State Management**
```
- Zustand 4.4         - Lightweight state management
- React Hot Toast 2.4 - Beautiful notifications
```

### **UI Components**
```
- Headless UI 1.7     - Unstyled, accessible components
- Heroicons 2.1       - Beautiful icons
- Recharts 2.10       - Charts and data viz
```

### **HTTP Client**
```
- Axios 1.6           - Promise-based HTTP client
- date-fns 2.30       - Date utilities
```

---

## 📦 **PROJECT STRUCTURE**

```
admin-dashboard/
├── src/
│   ├── components/         # Reusable UI components
│   │   ├── Layout.tsx      # Main layout with sidebar/header
│   │   ├── StatCard.tsx    # Statistics display cards
│   │   ├── DataTable.tsx   # Sortable data tables
│   │   ├── Modal.tsx       # Modal dialogs
│   │   └── ...
│   │
│   ├── pages/              # Page components
│   │   ├── Login.tsx       # Beautiful login page
│   │   ├── Dashboard.tsx   # Dashboard home
│   │   ├── Users.tsx       # User management
│   │   ├── Streams.tsx     # Stream management
│   │   ├── Categories.tsx  # Category management
│   │   ├── Analytics.tsx   # Analytics & charts
│   │   ├── Sessions.tsx    # Live sessions monitor
│   │   ├── Transcoding.tsx # Transcoding jobs
│   │   └── Settings.tsx    # Settings page
│   │
│   ├── services/           # API services
│   │   └── api.ts          # API client & endpoints
│   │
│   ├── stores/             # State management
│   │   └── authStore.ts    # Authentication store
│   │
│   ├── types/              # TypeScript types
│   │   └── index.ts        # Type definitions
│   │
│   ├── App.tsx             # Main app component
│   ├── main.tsx            # Entry point
│   └── index.css           # Global styles
│
├── public/                 # Static assets
├── index.html              # HTML template
├── package.json            # Dependencies
├── vite.config.ts          # Vite configuration
├── tailwind.config.js      # Tailwind configuration
├── tsconfig.json           # TypeScript config
└── README.md               # This file
```

---

## 🎨 **UI/UX DESIGN PRINCIPLES**

### **Color Palette**
```css
Primary:   #3b82f6 (Blue 500)
Success:   #10b981 (Green 500)
Warning:   #f59e0b (Amber 500)
Danger:    #ef4444 (Red 500)
Dark:      #0f172a (Slate 900)
Light:     #f8fafc (Slate 50)
```

### **Typography**
```
Headings:  Bold, gradient text
Body:      Regular, readable
Labels:    Medium weight, uppercase
```

### **Spacing**
```
Consistent 8px grid system
Generous padding and margins
Balanced white space
```

### **Animations**
```
Fade in:     Smooth page transitions
Slide up:    Content reveal
Hover:       Scale and shadow effects
Loading:     Shimmer and pulse
```

---

## 🔧 **SETUP & INSTALLATION**

### **1. Install Dependencies**
```bash
cd admin-dashboard
npm install
```

### **2. Environment Configuration**
Create `.env` file:
```bash
VITE_API_URL=http://localhost:80
```

### **3. Development Server**
```bash
npm run dev
```
Access: `http://localhost:3000`

### **4. Build for Production**
```bash
npm run build
```
Output: `dist/` directory

### **5. Preview Production Build**
```bash
npm run preview
```

---

## 🔐 **DEMO CREDENTIALS**

### **Admin Account**
```
Username: admin
Password: admin123
Access:   Full admin privileges
```

### **Test User**
```
Username: testuser
Password: admin123
Access:   Limited user access
```

---

## 📊 **DASHBOARD FEATURES**

### **1. Dashboard Home**
- Real-time user count
- Active streams monitor
- Revenue statistics
- Bandwidth usage charts
- Recent activity feed
- Quick action buttons

### **2. User Management**
- View all users (paginated)
- Search and filter users
- Create new users
- Edit user details
- Suspend/activate accounts
- Delete users (soft delete)
- User details modal
- Package assignment

### **3. Stream Management**
- List all streams
- Filter by type (Live/VOD)
- Filter by category
- Search streams
- Create new streams
- Edit stream details
- Upload stream sources
- Bitrate configurations
- Stream analytics

### **4. Categories**
- View all categories
- Create categories
- Edit category details
- Delete categories
- Icon management
- Stream count per category

### **5. Analytics**
- User growth charts
- Stream views over time
- Revenue trends
- Geographic distribution
- Popular content
- Peak usage times
- Device statistics
- Export reports (CSV/PDF)

### **6. Live Sessions**
- Real-time viewer list
- Connection details
- Bandwidth usage per session
- Session duration
- Terminate sessions
- Auto-refresh (every 5s)

### **7. Transcoding**
- View all jobs
- Job status (pending/processing/completed/failed)
- Progress bars
- Cancel running jobs
- Retry failed jobs
- Job details modal
- Create new jobs
- Preset management

### **8. Settings**
- General settings
- SMTP configuration
- CDN settings
- Storage configuration
- API keys management
- Rate limiting
- Maintenance mode
- Backup & restore

---

## 🎯 **API ENDPOINTS USED**

### **Authentication**
```
POST   /api/v1/auth/login       - User login
POST   /api/v1/auth/logout      - User logout
POST   /api/v1/auth/refresh     - Refresh token
GET    /api/v1/auth/me          - Get current user
PUT    /api/v1/auth/profile     - Update profile
POST   /api/v1/auth/password    - Change password
```

### **Users**
```
GET    /api/v1/admin/users      - List users
GET    /api/v1/admin/users/:id  - Get user
POST   /api/v1/admin/users      - Create user
PUT    /api/v1/admin/users/:id  - Update user
DELETE /api/v1/admin/users/:id  - Delete user
```

### **Streams**
```
GET    /api/v1/streams          - List streams
GET    /api/v1/streams/:id      - Get stream
GET    /api/v1/streams/search   - Search streams
POST   /api/v1/admin/streams    - Create stream
PUT    /api/v1/admin/streams/:id - Update stream
DELETE /api/v1/admin/streams/:id - Delete stream
```

### **Analytics**
```
GET    /api/v1/admin/analytics/dashboard  - Dashboard stats
GET    /api/v1/admin/analytics/streams    - Stream analytics
GET    /api/v1/admin/analytics/users      - User analytics
GET    /api/v1/admin/analytics/revenue    - Revenue data
```

### **Sessions**
```
GET    /api/v1/admin/sessions/active      - Active sessions
POST   /api/v1/admin/sessions/:id/terminate - Terminate session
```

---

## 🎨 **CUSTOM CSS UTILITIES**

### **Card Styles**
```css
.card           - White card with shadow
.glass          - Glassmorphism effect
.stat-card      - Animated stat card
```

### **Button Variants**
```css
.btn-primary    - Primary action button
.btn-secondary  - Secondary button
.btn-danger     - Destructive action
```

### **Badge Variants**
```css
.badge-success  - Green success badge
.badge-warning  - Yellow warning badge
.badge-danger   - Red danger badge
.badge-info     - Blue info badge
```

### **Input Styles**
```css
.input          - Styled input field
.table          - Styled data table
.skeleton       - Loading skeleton
```

---

## 📱 **RESPONSIVE BREAKPOINTS**

```
sm:   640px   - Mobile landscape
md:   768px   - Tablet
lg:   1024px  - Desktop
xl:   1280px  - Large desktop
2xl:  1536px  - Extra large
```

---

## ⚡ **PERFORMANCE OPTIMIZATIONS**

- ✅ Code splitting with React.lazy
- ✅ Tree shaking (Vite)
- ✅ Minification and compression
- ✅ Image optimization
- ✅ Lazy loading images
- ✅ Debounced search inputs
- ✅ Virtualized long lists
- ✅ Memoized components
- ✅ Cached API responses

---

## 🔒 **SECURITY FEATURES**

- ✅ JWT token authentication
- ✅ Secure token storage
- ✅ Automatic token refresh
- ✅ HTTPS enforcement (production)
- ✅ CORS protection
- ✅ XSS protection
- ✅ CSRF tokens
- ✅ Input sanitization
- ✅ SQL injection prevention
- ✅ Rate limiting

---

## 🐛 **ERROR HANDLING**

- ✅ Global error boundary
- ✅ API error interceptor
- ✅ Toast notifications
- ✅ Fallback UI
- ✅ Retry mechanisms
- ✅ Graceful degradation
- ✅ Error logging

---

## 📈 **FUTURE ENHANCEMENTS**

### **Phase 1 (Current)**
- [x] Authentication system
- [x] User management
- [x] Stream management
- [x] Basic analytics
- [x] Live sessions monitoring

### **Phase 2 (In Progress)**
- [ ] Advanced analytics (charts)
- [ ] Real-time notifications
- [ ] WebSocket integration
- [ ] Bulk operations
- [ ] Export functionality

### **Phase 3 (Planned)**
- [ ] Multi-language support (i18n)
- [ ] Theme customization
- [ ] Advanced filters
- [ ] Keyboard shortcuts
- [ ] Mobile app (React Native)

---

## 🎯 **BEST PRACTICES**

✅ **Component Organization**
- Small, focused components
- Reusable UI components
- Clear prop interfaces

✅ **State Management**
- Centralized auth store
- Local state for UI
- Optimistic updates

✅ **Code Quality**
- TypeScript for type safety
- ESLint for code style
- Prettier for formatting

✅ **Performance**
- Lazy loading
- Code splitting
- Optimized renders

---

## 🤝 **CONTRIBUTING**

This admin dashboard is part of the IPTV Platform Enterprise Edition. Follow these guidelines:

1. Use TypeScript for all new code
2. Follow the existing component structure
3. Add proper types and interfaces
4. Test on all screen sizes
5. Ensure dark mode compatibility
6. Add comments for complex logic

---

## 📞 **SUPPORT**

For issues or questions:
- Check documentation
- Review example code
- Contact development team

---

**Built with ❤️ for Enterprise IPTV Platform**

*Modern • Powerful • Beautiful*
