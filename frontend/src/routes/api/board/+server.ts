import type { Board } from '$lib/types/board';
import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';
import { isAdminSession, proxyBackend } from '$lib/server/backend';

export const POST: RequestHandler = async ({ request, fetch, cookies }) => {
	if (!isAdminSession(cookies)) {
		error(401, { message: 'Unauthorized' });
	}
	const body: Board = await request.json();
	body.alias = body.alias.replace('/', '');

	return proxyBackend(fetch, '/api/v1/board', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body),
	});
};
