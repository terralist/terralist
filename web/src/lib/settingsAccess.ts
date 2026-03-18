import { Auth } from '@/api/auth';

const settingsAccessCacheKey = '_auth.capabilities.settings';

const clearSettingsAccessCache = (): void => {
  sessionStorage.removeItem(settingsAccessCacheKey);
};

const canAccessSettings = async (): Promise<boolean> => {
  const cached = sessionStorage.getItem(settingsAccessCacheKey);
  if (cached != null) {
    return cached === 'true';
  }

  const { status, data } = await Auth.getSettingsCapability();
  const allowed = status === 'OK' && data.allowed === true;
  sessionStorage.setItem(settingsAccessCacheKey, String(allowed));
  return allowed;
};

export { canAccessSettings, clearSettingsAccessCache };
