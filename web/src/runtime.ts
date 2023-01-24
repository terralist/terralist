interface RuntimeEnv {
  TERRALIST_COMPANY_NAME?: string,
  TERRALIST_OAUTH_PROVIDERS?: string,
};

const runtimeEnv: RuntimeEnv = {
  TERRALIST_COMPANY_NAME: "{{.TERRALIST_COMPANY_NAME}}",
  TERRALIST_OAUTH_PROVIDERS: "{{.TERRALIST_OAUTH_PROVIDERS}}"
};

export default runtimeEnv;

export {
  type RuntimeEnv
};