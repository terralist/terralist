import axios from 'axios';
import { handleResponse, handleError } from '@/api/api.utils';
import config from '@/config';

type SessionDetails = {
  authorityId: string;
  name: string;
  email: string;
  groups: string[];
};

type SettingsCapability = {
  allowed: boolean;
};

const actions = {
  getSession: async () =>
    axios
      .get<SessionDetails>(config.runtime.TERRALIST_SESSION_ENDPOINT)
      .then(handleResponse<SessionDetails>)
      .catch(handleError),

  clearSession: async () =>
    axios
      .delete<boolean>(config.runtime.TERRALIST_SESSION_ENDPOINT)
      .then(handleResponse<boolean>)
      .catch(handleError),

  getSettingsCapability: async () =>
    axios
      .get<SettingsCapability>('/v1/api/auth/capabilities/settings')
      .then(handleResponse<SettingsCapability>)
      .catch(handleError)
};

const Auth = {
  getSession: async () => await actions.getSession(),
  clearSession: async () => await actions.clearSession(),
  getSettingsCapability: async () => await actions.getSettingsCapability()
};

export { Auth };
