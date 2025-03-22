import { Auth } from '@/api/auth';

type Session = {
  [k: string]: string;
};

type UserSession = {
  userName: string;
  userEmail: string;
};

const sessionKeys: Session = {
  userName: 'user.name',
  userEmail: 'user.email',
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
    Object.entries(session).forEach(([key, value]) =>
      sessionStorage.setItem(`_auth.session.${sessionKeys[key]}`, value)
    );
  },

  reset: () => {
    Object.values(sessionKeys).forEach(value =>
      sessionStorage.removeItem(`_auth.session.${value}`)
    );
  }
};

const isAvailable = (): boolean => {
  const session = actions.download();

  const isSessionSet = Object.values(session).every(v => v);

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
      userEmail: session.userEmail
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

      const session = {
        expireAt: expireAt.toISOString(),
        userName: data.name,
        userEmail: data.email
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
