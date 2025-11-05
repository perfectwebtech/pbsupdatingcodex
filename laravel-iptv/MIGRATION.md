# 📦 Migration Guide: Legacy to Laravel Platform

This guide helps you migrate from the old PHP platform to the new Laravel system.

## Migration Strategy

We recommend a **gradual migration** approach:

1. **Phase 1**: Run both systems in parallel (1-2 weeks)
2. **Phase 2**: Migrate data incrementally
3. **Phase 3**: Switch traffic gradually
4. **Phase 4**: Decommission old system

## Pre-Migration Checklist

- [ ] Backup current database
- [ ] Document current infrastructure
- [ ] Test new system with sample data
- [ ] Prepare rollback plan
- [ ] Inform users of upcoming changes

## Phase 1: Parallel Setup

### Step 1: Deploy New System

```bash
# Deploy Laravel platform on separate server/port
cd /home/user/pbsupdatingcodex/laravel-iptv
docker-compose up -d
```

### Step 2: Configure Dual Database Access

Update `.env` to access both databases:

```env
# New database
DB_CONNECTION=pgsql
DB_HOST=new-server.com
DB_DATABASE=iptv_platform

# Legacy database for migration
LEGACY_DB_CONNECTION=mysql
LEGACY_DB_HOST=old-server.com
LEGACY_DB_DATABASE=xtream_codes
```

### Step 3: Set Up Load Balancer

```nginx
# nginx.conf
upstream old_system {
    server old-server:8080;
}

upstream new_system {
    server new-server:8000;
}

# Route by user agent or parameter
map $http_user_agent $backend {
    default old_system;
    ~*NewApp new_system;
}

server {
    location / {
        proxy_pass http://$backend;
    }
}
```

## Phase 2: Data Migration

### Step 1: Create Migration Commands

```bash
# We'll create these artisan commands
docker-compose exec app php artisan make:command MigrateUsers
docker-compose exec app php artisan make:command MigrateStreams
docker-compose exec app php artisan make:command MigratePackages
```

### Step 2: Migrate Reference Data First

```bash
# 1. Migrate categories
docker-compose exec app php artisan migrate:legacy-categories

# 2. Migrate servers
docker-compose exec app php artisan migrate:legacy-servers

# 3. Migrate packages
docker-compose exec app php artisan migrate:legacy-packages

# 4. Migrate streams
docker-compose exec app php artisan migrate:legacy-streams
```

### Step 3: Migrate User Data in Batches

```bash
# Test with small batch first
docker-compose exec app php artisan migrate:legacy-users --limit=10

# Check data
docker-compose exec app php artisan tinker
>>> App\Models\User::count();

# Migrate all users
docker-compose exec app php artisan migrate:legacy-users --batch=1000
```

### Example Migration Command

```php
<?php
// app/Console/Commands/MigrateUsers.php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;
use App\Models\User;
use App\Models\Package;

class MigrateUsers extends Command
{
    protected $signature = 'migrate:legacy-users {--limit=100} {--batch=1000}';
    protected $description = 'Migrate users from legacy database';

    public function handle()
    {
        $legacyDb = DB::connection('legacy_mysql');
        $limit = $this->option('limit');
        $batch = $this->option('batch');

        $this->info("Migrating users from legacy database...");

        // Get users from old database
        $oldUsers = $legacyDb->table('users')
            ->limit($limit)
            ->get();

        $bar = $this->output->createProgressBar($oldUsers->count());

        foreach ($oldUsers as $oldUser) {
            // Map old user to new structure
            $newUser = User::updateOrCreate(
                ['username' => $oldUser->username],
                [
                    'password' => $oldUser->password, // Already hashed
                    'max_connections' => $oldUser->max_connections,
                    'is_trial' => (bool)$oldUser->is_trial,
                    'expires_at' => $oldUser->exp_date ?
                        date('Y-m-d H:i:s', $oldUser->exp_date) : null,
                    'is_active' => (bool)$oldUser->enabled,
                    'admin_enabled' => (bool)$oldUser->admin_enabled,
                    'created_at' => $oldUser->created_at,
                ]
            );

            $bar->advance();
        }

        $bar->finish();
        $this->newLine();
        $this->info("Migration completed successfully!");
    }
}
```

## Phase 3: Traffic Switching

### Gradual Rollout

```nginx
# Split traffic 10% new, 90% old
upstream backend {
    server old-server:8080 weight=90;
    server new-server:8000 weight=10;
}
```

**Week 1**: 10% new, 90% old
**Week 2**: 25% new, 75% old
**Week 3**: 50% new, 50% old
**Week 4**: 75% new, 25% old
**Week 5**: 100% new

### Monitor Metrics

```bash
# Monitor error rates
tail -f storage/logs/laravel.log | grep ERROR

# Monitor performance
docker-compose exec app php artisan monitor:performance

# Check active sessions
docker-compose exec app php artisan tinker
>>> App\Models\UserSession::active()->count();
```

## Phase 4: Decommission Legacy

### Before Decommissioning

- [ ] All users migrated
- [ ] All streams working
- [ ] No errors in logs for 7 days
- [ ] Performance metrics acceptable
- [ ] Backup legacy database
- [ ] Document final state

### Decommission Steps

```bash
# 1. Stop accepting new connections on old system
# Update nginx to return 503 for old endpoints

# 2. Wait for active sessions to complete
# Monitor legacy database for active_connections

# 3. Export final data
mysqldump -u root -p xtream_codes > legacy_final_backup.sql

# 4. Archive old system
tar -czf legacy_system_archive.tar.gz /home/xtreamcodes/

# 5. Shutdown old services
systemctl stop nginx
systemctl stop php-fpm
```

## Data Mapping Reference

### Users Table

| Legacy | New | Notes |
|--------|-----|-------|
| `id` | `id` | Keep same ID |
| `username` | `username` | |
| `password` | `password` | Already hashed |
| `exp_date` | `expires_at` | Convert Unix to datetime |
| `max_connections` | `max_connections` | |
| `enabled` | `is_active` | |
| `admin_enabled` | `admin_enabled` | |
| `is_trial` | `is_trial` | |
| `bouquet` | `package_id` | Map to packages |

### Streams Table

| Legacy | New | Notes |
|--------|-----|-------|
| `id` | `id` | |
| `stream_display_name` | `name` | |
| `type` | `type` | Map: 1=live, 2=vod, etc |
| `category_id` | `category_id` | |
| `stream_source` | `source_urls` | Parse JSON |
| `stream_icon` | `icon_url` | |
| `tv_archive_duration` | `archive_duration_days` | |

## Rollback Plan

If issues arise:

```bash
# 1. Switch nginx back to old system
# Update upstream to point to legacy

# 2. Stop new system
docker-compose down

# 3. Restore old database if needed
mysql -u root -p xtream_codes < backup.sql

# 4. Investigate issues
# Check logs, fix problems

# 5. Retry migration
# Address root causes before retry
```

## Validation Checklist

After migration:

- [ ] All users can login
- [ ] Streams play correctly
- [ ] HLS segments deliver properly
- [ ] Connection limits work
- [ ] Geographic blocking works
- [ ] Package permissions correct
- [ ] EPG data displays
- [ ] Archive/DVR functional
- [ ] Admin panel accessible
- [ ] Logs are clean

## Performance Benchmarks

Compare before/after:

```bash
# API response time
ab -n 1000 -c 10 http://localhost:8000/api/v1/streams

# Stream startup time
time curl http://localhost:8000/api/v1/stream/1/playlist.m3u8

# Database query time
docker-compose exec app php artisan db:benchmark
```

## Support During Migration

- Monitor Slack channel: #migration-support
- Daily standups at 9 AM
- Emergency contact: on-call engineer
- Rollback authority: CTO approval required

## Post-Migration

1. **Monitor for 30 days**
2. **Optimize based on metrics**
3. **Archive legacy system**
4. **Document lessons learned**
5. **Celebrate success! 🎉**
