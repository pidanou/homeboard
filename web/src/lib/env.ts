// PUBLIC_ENV=local (default) or production
const appEnv = import.meta.env.PUBLIC_ENV ?? 'local';

export const isLocal = appEnv === 'local';
export const isSaaS = appEnv === 'production';

/**
 * Feature flags by environment:
 *
 * LOCAL only:
 *   - (none remaining — setup page removed, API URL configured via PUBLIC_API_URL)
 *
 * SAAS only:
 *   - "New household" button (sidebar + home page)
 */
