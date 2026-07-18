import { cookieSecure } from '$lib/auth';
import { backendUrl } from '$lib/server/backend';
import type { Cookies } from '@sveltejs/kit';

export const FOLLOWED_COOKIE = 'followed_threads';
export const FOLLOWED_MAX = 50;

export type FollowedEntry = { lastSeenReplies: number };
/** Insertion-ordered map of thread id → last-seen reply count. */
export type FollowedMap = Record<string, FollowedEntry>;

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

export function parseFollowedCookie(raw: string | undefined): FollowedMap {
	if (!raw) return {};
	try {
		const parsed = JSON.parse(raw) as unknown;
		if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {};
		const out: FollowedMap = {};
		for (const [id, value] of Object.entries(parsed as Record<string, unknown>)) {
			if (!id || typeof value !== 'object' || value === null) continue;
			const lastSeen = Number((value as FollowedEntry).lastSeenReplies);
			out[id] = { lastSeenReplies: Number.isFinite(lastSeen) ? lastSeen : 0 };
		}
		return out;
	} catch {
		return {};
	}
}

export function serializeFollowedCookie(map: FollowedMap): string {
	return JSON.stringify(map);
}

export function setFollowedCookie(cookies: Cookies, map: FollowedMap) {
	cookies.set(FOLLOWED_COOKIE, serializeFollowedCookie(map), {
		path: '/',
		httpOnly: false,
		sameSite: 'lax',
		secure: cookieSecure,
		maxAge: 60 * 60 * 24 * 30,
	});
}

/** Add or update a follow; drops oldest insertion-order entries past FOLLOWED_MAX. */
export function addFollow(map: FollowedMap, id: string, lastSeenReplies: number): FollowedMap {
	const next: FollowedMap = { ...map };
	if (id in next) {
		next[id] = { lastSeenReplies };
		return next;
	}
	const keys = Object.keys(next);
	while (keys.length >= FOLLOWED_MAX) {
		const oldest = keys.shift();
		if (!oldest) break;
		delete next[oldest];
	}
	next[id] = { lastSeenReplies };
	return next;
}

export function removeFollow(map: FollowedMap, id: string): FollowedMap {
	const next: FollowedMap = { ...map };
	delete next[id];
	return next;
}

/** Update lastSeenReplies if the thread is followed. Returns null when unchanged / not followed. */
export function markSeen(map: FollowedMap, id: string, replies: number): FollowedMap | null {
	const entry = map[id];
	if (!entry) return null;
	if (entry.lastSeenReplies === replies) return null;
	return { ...map, [id]: { lastSeenReplies: replies } };
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
