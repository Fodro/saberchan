import { backendUrl } from '$lib/server/backend';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, getClientAddress }) => {
	return fetch(backendUrl('/api/v1/captcha'), {
		headers: { 'X-Forwarded-For': getClientAddress() },
	});
};
