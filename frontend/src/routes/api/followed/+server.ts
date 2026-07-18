import { json } from '@sveltejs/kit';
import { loadFollowedSummary } from '$lib/server/followed';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ fetch, cookies }) => {
	const summary = await loadFollowedSummary(fetch, cookies);
	return json({
		threads: summary.threads,
		dirty_count: summary.dirty_count,
	});
};
