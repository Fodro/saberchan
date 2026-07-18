import { describe, expect, it } from 'vitest';
import {
	FOLLOWED_MAX,
	addFollow,
	markSeen,
	parseFollowedCookie,
	removeFollow,
	serializeFollowedCookie,
	type FollowedMap,
} from './followedMap';

describe('parseFollowedCookie / serializeFollowedCookie', () => {
	it('returns empty for missing / invalid', () => {
		expect(parseFollowedCookie(undefined)).toEqual({});
		expect(parseFollowedCookie('')).toEqual({});
		expect(parseFollowedCookie('[]')).toEqual({});
		expect(parseFollowedCookie('not-json')).toEqual({});
	});

	it('round-trips a map and coerces lastSeen', () => {
		const map = { a: { lastSeenReplies: 3 }, b: { lastSeenReplies: Number.NaN } };
		const parsed = parseFollowedCookie(serializeFollowedCookie(map));
		expect(parsed.a.lastSeenReplies).toBe(3);
		expect(parsed.b.lastSeenReplies).toBe(0);
	});
});

describe('addFollow / removeFollow / markSeen', () => {
	it('adds and updates entries', () => {
		let map = addFollow({}, 't1', 0);
		expect(map.t1.lastSeenReplies).toBe(0);
		map = addFollow(map, 't1', 5);
		expect(map.t1.lastSeenReplies).toBe(5);
	});

	it('evicts oldest when over FOLLOWED_MAX', () => {
		let map: FollowedMap = {};
		for (let i = 0; i < FOLLOWED_MAX; i++) {
			map = addFollow(map, `id${i}`, i);
		}
		expect(Object.keys(map)).toHaveLength(FOLLOWED_MAX);
		map = addFollow(map, 'newest', 99);
		expect(Object.keys(map)).toHaveLength(FOLLOWED_MAX);
		expect(map).not.toHaveProperty('id0');
		expect(map.newest.lastSeenReplies).toBe(99);
	});

	it('removes and marks seen', () => {
		const map = addFollow({ t1: { lastSeenReplies: 1 } }, 't2', 2);
		expect(removeFollow(map, 't1')).toEqual({ t2: { lastSeenReplies: 2 } });
		expect(markSeen(map, 'missing', 9)).toBeNull();
		expect(markSeen(map, 't2', 2)).toBeNull();
		expect(markSeen(map, 't2', 4)).toEqual({
			t1: { lastSeenReplies: 1 },
			t2: { lastSeenReplies: 4 },
		});
	});
});
