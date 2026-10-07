import type { ApiClient } from './client';
import type { components } from './schema';

export type Meta = components['schemas']['Meta'];

export function getMeta(client: ApiClient): Promise<Meta> {
	return client.get<Meta>('/api/v1/meta');
}
