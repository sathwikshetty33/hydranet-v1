import { api } from './client';

const ENDPOINT = '/api/v1/incidents';

export async function listIncidents() {
  const data = await api.get(ENDPOINT);
  return Array.isArray(data) ? data : [];
}
