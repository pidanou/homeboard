// PUBLIC_ENV=local (default) or production
const appEnv = import.meta.env.PUBLIC_ENV ?? 'local';

export const isLocal = appEnv === 'local';
export const isSaaS = appEnv === 'production';

/**
 * Feature flags by environment:
 *
 * LOCAL only:
 *   - "Change server" button on the login page
 *   - /setup page (configure API URL for Capacitor native builds)
 *   - configurable API URL via localStorage
 *
 * SAAS only:
 *   - (none yet — add billing, plan management, etc. here)
 */
