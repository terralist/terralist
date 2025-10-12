import { Auth } from '@/api/auth';
import { defaultIfNull } from './utils';

type Session = {
  userName: string;
  userEmail: string;
  userGroups: string;
  expireAt: string;
};

type UserSession = {
  userName: string;
  userEmail: string;
  userGroups: string[];
};

const sessionKeys: Record<keyof Session, string> = {
  userName: 'user.name',
  userEmail: 'user.email',
  userGroups: 'user.groups',
  expireAt: 'expire_at'
};

const actions = {
  download: (): Session => {
    return Object.fromEntries(
      Object.entries(sessionKeys)
        .map(([key, value]) => [
          key,
          sessionStorage.getItem(`_auth.session.${value}`)
        ])
        .filter(([, value]) => value != null)
    );
  },

  upload: (session: Session) => {
    (Object.keys(session) as (keyof Session)[]).forEach(key => {
      const value = session[key];
      sessionStorage.setItem(
        `_auth.session.${sessionKeys[key]}`,
        String(value)
      );
    });
  },

  reset: () => {
    Object.values(sessionKeys).forEach(value =>
      sessionStorage.removeItem(`_auth.session.${value}`)
    );
  }
};

const isAvailable = (): boolean => {
  const session = actions.download();

  const isSessionSet = Object.values(session).every(
    v => v != undefined && v != null
  );

  if (!isSessionSet) {
    return false;
  }

  if (session?.expireAt) {
    if (new Date(session.expireAt).getTime() <= new Date().getTime()) {
      return false;
    }
  } else {
    return false;
  }

  return true;
};

const UserStore = {
  isAvailable: () => isAvailable(),

  get: (): UserSession | null => {
    if (!isAvailable()) {
      return null;
    }

    const session = actions.download();

    return {
      userName: session.userName,
      userEmail: session.userEmail,
      userGroups: session.userGroups.split('#').filter(g => g.length > 0)
    } satisfies UserSession;
  },

  refresh: async () => {
    const { data, status } = await Auth.getSession();

    if (status == 'OK') {
      const SESSION_EXPIRE_AFTER_MINUTES = 1;

      const expireAt = new Date();
      expireAt.setTime(
        new Date().getTime() + SESSION_EXPIRE_AFTER_MINUTES * 60 * 1000
      );

      const session: Session = {
        expireAt: expireAt.toISOString(),
        userName: data.name,
        userEmail: data.email,
        userGroups: defaultIfNull(data.groups, []).join('#')
      };

      actions.upload(session);
    }
  },

  clear: async () => {
    const { status } = await Auth.clearSession();

    if (status === 'OK') {
      actions.reset();
    }
  }
};

export { UserStore };
