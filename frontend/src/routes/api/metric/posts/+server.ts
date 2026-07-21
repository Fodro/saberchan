import { adminBackendHeaders, proxyBackend } from '$lib/server/backend';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, cookies, url }) => {
	return proxyBackend(fetch, `/api/v1/metric/posts${url.search}`, {
		headers: await adminBackendHeaders(cookies),
	});
};