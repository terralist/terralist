type Artifact = {
  id: string,
  fullName: string,
  authority: string,
  name: string,
  provider?: string,
  type: "provider" | "module",
};

const cache: {
  artifacts: Artifact[],
} = {
  artifacts: [],
};

const fetchArtifacts = (refresh: boolean = false) => {
  if (refresh) {
    ;
  }

  cache.artifacts = [
    { id: "1", fullName: "HashiCorp/aws", authority: 'HashiCorp', name: 'aws', type: 'provider' },
    { id: "2", fullName: "HashiCorp/null", authority: 'HashiCorp', name: 'null', type: 'provider' },
    { id: "3", fullName: "HashiCorp/vpc/aws", authority: 'HashiCorp', name: 'vpc', provider: 'aws', type: 'module' },
    { id: "4", fullName: "HashiCorp/iam/aws", authority: 'HashiCorp', name: 'iam', provider: 'aws', type: 'module' },
    { id: "5", fullName: "Heroku/heroku", authority: 'Heroku', name: 'heroku', type: 'provider' },
    { id: "6", fullName: "Heroku/heroku2", authority: 'Heroku', name: 'heroku2', type: 'provider' },
    { id: "7", fullName: "Heroku/heroku3", authority: 'Heroku', name: 'heroku3', type: 'provider' },
    { id: "8", fullName: "Heroku/heroku4", authority: 'Heroku', name: 'heroku4', type: 'provider' },
    { id: "9", fullName: "Heroku/heroku5", authority: 'Heroku', name: 'heroku5', type: 'provider' },
    { id: "10", fullName: "Heroku/heroku6", authority: 'Heroku', name: 'heroku6', type: 'provider' },
    { id: "11", fullName: "Heroku/heroku7", authority: 'Heroku', name: 'heroku7', type: 'provider' },
  ];

  return cache.artifacts;
};

export type { Artifact };

export {
  fetchArtifacts
};