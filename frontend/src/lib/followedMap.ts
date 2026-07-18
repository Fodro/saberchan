export const FOLLOWED_MAX = 50;

export type FollowedEntry = { lastSeenReplies: number };
/** Insertion-ordered map of thread id → last-seen reply count. */
export type FollowedMap = Record<string, FollowedEntry>;

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
