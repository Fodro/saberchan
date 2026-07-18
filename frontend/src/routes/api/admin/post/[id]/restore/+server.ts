import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { adminBackendHeaders, isAdminSession, proxyBackend } from '$lib/server/backend';

export const POST: RequestHandler = async ({ params, fetch, cookies }) => {
	if (!isAdminSession(cookies)) {
		error(401, { message: 'Unauthorized' });
	}
	const id = params.id;
	if (!id) error(400, { message: 'id is required' });

	return proxyBackend(fetch, `/api/v1/post/${id}/restore`, {
		method: 'POST',
		headers: adminBackendHeaders(cookies),
	});
};
