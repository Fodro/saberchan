import { cookieSecure } from '$lib/auth';
import { backendUrl } from '$lib/server/backend';
import {
	FOLLOWED_MAX,
	addFollow,
	markSeen,
	parseFollowedCookie,
	removeFollow,
	serializeFollowedCookie,
	type FollowedEntry,
	type FollowedMap,
} from '$lib/followedMap';
import type { Cookies } from '@sveltejs/kit';

export const FOLLOWED_COOKIE = 'followed_threads';
export { FOLLOWED_MAX, addFollow, markSeen, parseFollowedCookie, removeFollow, serializeFollowedCookie };
export type { FollowedEntry, FollowedMap };

export type BackendFollowStatus = {
	id: string;
	title: string;
	board_alias: string;
	replies_count: number;
	dead: boolean;
};

export type FollowedThreadSummary = {
	id: string;
	title: string;
	board_alias: string;
	new_posts: number;
	dead: boolean;
	href: string;
};

export type FollowedSummary = {
	threads: FollowedThreadSummary[];
	/** Sum of new_posts on non-dead threads (navbar badge). */
	dirty_count: number;
	/** Cookie order of followed thread ids. */
	ids: string[];
};

export function setFollowedCookie(cookies: Cookies, map: FollowedMap) {
	cookies.set(FOLLOWED_COOKIE, serializeFollowedCookie(map), {
		path: '/',
		httpOnly: false,
		sameSite: 'lax',
		secure: cookieSecure,
		maxAge: 60 * 60 * 24 * 30,
	});
}

export async function loadFollowedSummary(
	fetchFn: typeof fetch,
	cookies: Cookies,
): Promise<FollowedSummary> {
	const map = parseFollowedCookie(cookies.get(FOLLOWED_COOKIE));
	const ids = Object.keys(map);
	if (ids.length === 0) {
		return { threads: [], dirty_count: 0, ids: [] };
	}

	let statuses: BackendFollowStatus[] = [];
	try {
		const qs = ids.map(encodeURIComponent).join(',');
		const res = await fetchFn(backendUrl(`/api/v1/follow?ids=${qs}`));
		if (res.ok) {
			const json = (await res.json()) as unknown;
			if (Array.isArray(json)) {
				statuses = json as BackendFollowStatus[];
			}
		}
	} catch {
		// Backend unavailable — badge stays 0 until /api/v1/follow exists.
	}

	const byId = new Map(statuses.map((s) => [String(s.id), s]));
	const threads: FollowedThreadSummary[] = [];
	let dirty_count = 0;

	for (const id of ids) {
		const status = byId.get(id);
		if (!status) continue;
		const lastSeen = map[id]?.lastSeenReplies ?? 0;
		const replies = Number(status.replies_count) || 0;
		const dead = Boolean(status.dead);
		const new_posts = dead ? 0 : Math.max(0, replies - lastSeen);
		if (!dead) dirty_count += new_posts;
		threads.push({
			id: String(status.id),
			title: status.title || id,
			board_alias: status.board_alias || '',
			new_posts,
			dead,
			href: `/board/${status.board_alias}/thread/${status.id}`,
		});
	}

	return { threads, dirty_count, ids };
}
