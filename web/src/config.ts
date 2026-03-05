type RuntimeVariables = {
  TERRALIST_HOST_URL: string;
  TERRALIST_CANONICAL_DOMAIN: string;
  TERRALIST_COMPANY_NAME: string;
  TERRALIST_OAUTH_PROVIDERS: string[];
  TERRALIST_AUTHORIZATION_ENDPOINT: string;
  TERRALIST_SESSION_ENDPOINT: string;
  TERRALIST_AUTHORIZED_USERS: string;
  TERRALIST_SAML_DISPLAY_NAME: string;
};

type BuildVariables = {
  TERRALIST_VERSION: string;
};

const DEFAULT_RUNTIME_VARIABLES: RuntimeVariables = {
  TERRALIST_HOST_URL: 'http://localhost:5758',
  TERRALIST_CANONICAL_DOMAIN: 'localhost',
  TERRALIST_COMPANY_NAME: '',
  TERRALIST_OAUTH_PROVIDERS: ['github', 'bitbucket', 'gitlab', 'saml'],
  // TODO: These should point to a mock endpoint for local development
  TERRALIST_AUTHORIZATION_ENDPOINT: '',
  TERRALIST_SESSION_ENDPOINT: '',
  TERRALIST_AUTHORIZED_USERS: '',
  TERRALIST_SAML_DISPLAY_NAME: 'SSO'
};

class Configuration {
  runtime: RuntimeVariables;
  build: BuildVariables;

  constructor() {
    this.runtime = DEFAULT_RUNTIME_VARIABLES;

    this.build = {
      TERRALIST_VERSION: import.meta.env.TERRALIST_VERSION || 'dev'
    };
  }

  async refresh(): Promise<void> {
    const cache = sessionStorage.getItem('runtime');
    if (cache) {
      this.runtime = JSON.parse(cache);
      return;
    }

    this.runtime = DEFAULT_RUNTIME_VARIABLES;

    const resp = await fetch('/internal/runtime.json');

    if (resp.ok && resp.status === 200) {
      const data = await resp.json();

      this.runtime.TERRALIST_HOST_URL = data['host'];
      this.runtime.TERRALIST_CANONICAL_DOMAIN = data['domain'];
      this.runtime.TERRALIST_COMPANY_NAME = data['company'];
      this.runtime.TERRALIST_AUTHORIZED_USERS = data['authorized_users'];
      this.runtime.TERRALIST_SAML_DISPLAY_NAME =
        data['saml_display_name'] || 'SSO';
      this.runtime.TERRALIST_OAUTH_PROVIDERS = data['auth']['providers'];
      this.runtime.TERRALIST_AUTHORIZATION_ENDPOINT = data['auth']['endpoint'];
      this.runtime.TERRALIST_SESSION_ENDPOINT =
        data['auth']['session_endpoint'];
    }

    sessionStorage.setItem('runtime', JSON.stringify(this.runtime));
  }
}

const config: Configuration = new Configuration();

export default config;

export { type Configuration };
