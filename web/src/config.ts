interface RuntimeVariables {
  TERRALIST_HOST_URL?: string,
  TERRALIST_CANONICAL_DOMAIN?: string,
  TERRALIST_COMPANY_NAME?: string,
  TERRALIST_OAUTH_PROVIDERS?: string[],
  TERRALIST_AUTHORIZATION_ENDPOINT?: string,
  TERRALIST_SESSION_ENDPOINT?: string,
}

interface BuildVariables {
  TERRALIST_VERSION: string,
}

class Configuration {
  runtime: RuntimeVariables;
  build: BuildVariables;

  constructor() {
    this.build = {
      TERRALIST_VERSION: import.meta.env.TERRALIST_VERSION || "dev",
    };
  }

  async refresh() {
    let cache = sessionStorage.getItem("runtime");
    if (cache) {
      this.runtime = JSON.parse(cache);
      return;
    }

    this.runtime = {} satisfies RuntimeVariables;

    let resp = await fetch("/internal/runtime.json");
    
    if (resp.ok && resp.status === 200) {
      let data = await resp.json();

      this.runtime.TERRALIST_HOST_URL = data["host"];
      this.runtime.TERRALIST_CANONICAL_DOMAIN = data["domain"];
      this.runtime.TERRALIST_COMPANY_NAME = data["company"];
      this.runtime.TERRALIST_OAUTH_PROVIDERS = data["auth"]["providers"];
      this.runtime.TERRALIST_AUTHORIZATION_ENDPOINT = data["auth"]["endpoint"];
      this.runtime.TERRALIST_SESSION_ENDPOINT = data["auth"]["session_endpoint"];
    } else {
      this.runtime.TERRALIST_HOST_URL = "http://localhost:5758";
      this.runtime.TERRALIST_CANONICAL_DOMAIN = "localhost";
      this.runtime.TERRALIST_COMPANY_NAME = "";
      this.runtime.TERRALIST_OAUTH_PROVIDERS = ["github", "bitbucket", "gitlab", "google"];
      this.runtime.TERRALIST_AUTHORIZATION_ENDPOINT = ""; // TODO: This should point to a mock endpoint for local development
      this.runtime.TERRALIST_SESSION_ENDPOINT = ""; // TODO: This should point to a mock endpoint for local development 
    }

    sessionStorage.setItem("runtime", JSON.stringify(this.runtime));
  }
}

const config: Configuration = new Configuration();

export default config;

export {
  type Configuration
};