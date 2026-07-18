import { error, json } from '@sveltejs/kit';
import { backendUrl } from '$lib/server/backend';
import {
	FOLLOWED_COOKIE,
	addFollow,
	parseFollowedCookie,
	removeFollow,
	setFollowedCookie,
} from '$lib/server/followed';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ params, request, cookies, fetch }) => {
	const id = params.id?.trim();
	if (!id) {
		error(400, { message: 'thread id is required' });
	}

	let lastSeenReplies = 0;
	try {
		const body = (await request.json()) as { lastSeenReplies?: number };
		if (typeof body?.lastSeenReplies === 'number' && Number.isFinite(body.lastSeenReplies)) {
			lastSeenReplies = Math.max(0, Math.floor(body.lastSeenReplies));
		}
	} catch {
		// Optional body.
	}

	let be: Response;
	try {
		be = await fetch(backendUrl(`/api/v1/follow/${encodeURIComponent(id)}`), {
			method: 'POST',
		});
	} catch {
		error(502, { message: 'follow backend unreachable' });
	}
	if (!be.ok) {
		const msg = (await be.text()) || `follow failed (${be.status})`;
		error(be.status, { message: msg });
	}

	const map = addFollow(parseFollowedCookie(cookies.get(FOLLOWED_COOKIE)), id, lastSeenReplies);
	setFollowedCookie(cookies, map);

	return json({ ok: true });
};

export const DELETE: RequestHandler = async ({ params, cookies }) => {
	const id = params.id?.trim();
	if (!id) {
		error(400, { message: 'thread id is required' });
	}

	const map = removeFollow(parseFollowedCookie(cookies.get(FOLLOWED_COOKIE)), id);
	setFollowedCookie(cookies, map);

	return json({ ok: true });
};
