import runtimeEnv, { type RuntimeEnv } from "./runtime";

interface Configuration {
  runtime: {
    env: RuntimeEnv
  },
  build: {
    env: {
      readonly TERRALIST_VERSION: string,
    }
  }
}

const goTplPattern = new RegExp("^[{]{2}\s*[.][A-Z_]+\s*[}]{2}$");

const orDefault = (value: string, def: string = "") => {
  return goTplPattern.test(value) ? def : value;
};

const config: Configuration = {
  runtime: {
    env: {
      TERRALIST_CANONICAL_DOMAIN: orDefault(runtimeEnv.TERRALIST_CANONICAL_DOMAIN, "localhost"),
      TERRALIST_COMPANY_NAME: orDefault(runtimeEnv.TERRALIST_COMPANY_NAME, ""),
      TERRALIST_OAUTH_PROVIDERS: orDefault(runtimeEnv.TERRALIST_OAUTH_PROVIDERS, JSON.stringify(["github", "google", "bitbucket", "gitlab"])),
    },
  },
  build: {
    env: {
      TERRALIST_VERSION: import.meta.env.TERRALIST_VERSION || "dev",
    },
  }
} satisfies Configuration;

export default config;

export {
  type Configuration
};