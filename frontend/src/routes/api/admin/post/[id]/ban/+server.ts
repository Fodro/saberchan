import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { adminBackendHeaders, isAdminSession, proxyBackend } from '$lib/server/backend';

export const POST: RequestHandler = async ({ params, request, fetch, cookies }) => {
	if (!(await isAdminSession(cookies))) {
		error(401, { message: 'Unauthorized' });
	}
	const id = params.id;
	if (!id) error(400, { message: 'id is required' });

	const body = await request.json();

	return proxyBackend(fetch, `/api/v1/post/${id}/ban`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			...(await adminBackendHeaders(cookies)),
		},
		body: JSON.stringify(body),
	});
};
