import { api } from './client';

const ENDPOINT = '/api/v1/agents';

export async function listAgents() {
  const data = await api.get(ENDPOINT);
  return Array.isArray(data) ? data : [];
}

// Returns the created agent including auth_token. The token is in the
// response exactly once; the caller is responsible for surfacing it to the
// user and discarding it after.
export async function createAgent({ name, clusterName }) {
  return api.post(ENDPOINT, {
    name,
    cluster_name: clusterName,
  });
}
