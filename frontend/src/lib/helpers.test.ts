import { afterEach, describe, expect, it, vi } from 'vitest';
import {
	base64ToArrayBuffer,
	bufferToBase64,
	trimLargeWords,
	verifyExp,
} from './helpers';
import { composeErrorMessageFactory, formatBannedMessage, parsePossiblyWrappedJson } from './compose';

describe('verifyExp', () => {
	afterEach(() => {
		vi.useRealTimers();
	});

	it('treats missing exp as expired/invalid session', () => {
		expect(verifyExp(undefined)).toBe(true);
	});

	it('compares unix seconds to now', () => {
		vi.useFakeTimers();
		vi.setSystemTime(new Date('2020-01-01T00:00:00Z'));
		const nowSec = Math.floor(Date.now() / 1000);
		expect(verifyExp(nowSec + 60)).toBe(false);
		expect(verifyExp(nowSec - 1)).toBe(true);
	});
});

describe('trimLargeWords', () => {
	it('inserts breaks into long tokens', () => {
		const word = 'a'.repeat(60);
		const out = trimLargeWords(`hi ${word} bye`);
		expect(out.includes('\n')).toBe(true);
		expect(out.startsWith('hi ')).toBe(true);
		expect(out.endsWith(' bye')).toBe(true);
	});

	it('leaves short words alone', () => {
		expect(trimLargeWords('short words only')).toBe('short words only');
	});
});

describe('bufferToBase64 / base64ToArrayBuffer', () => {
	it('round-trips ArrayBuffer', () => {
		const bytes = new Uint8Array([1, 2, 255, 0]);
		const b64 = bufferToBase64(bytes.buffer);
		expect(b64.length).toBeGreaterThan(0);
		expect(new Uint8Array(base64ToArrayBuffer(b64))).toEqual(bytes);
	});

	it('handles empty / string inputs', () => {
		expect(bufferToBase64(null)).toBe('');
		expect(bufferToBase64('already')).toBe('already');
	});
});

describe('composeErrorMessageFactory', () => {
	it('maps codes through t()', () => {
		const msg = composeErrorMessageFactory((k) => `T:${k}`);
		expect(msg('title_required')).toBe('T:common.compose.title_required');
		expect(msg('file_type')).toBe('T:common.file.limitType');
	});
});

describe('ban JSON helpers', () => {
	it('parses wrapped BFF message bodies', () => {
		const inner = JSON.stringify({ code: 'banned', reason: 'spam', until: '2099-01-01T00:00:00Z' });
		expect(parsePossiblyWrappedJson(JSON.stringify({ message: inner }))).toEqual({
			code: 'banned',
			reason: 'spam',
			until: '2099-01-01T00:00:00Z',
		});
	});

	it('parses flat ban payloads and rejects garbage', () => {
		expect(parsePossiblyWrappedJson('{"code":"banned","reason":"x"}')).toMatchObject({
			code: 'banned',
			reason: 'x',
		});
		expect(parsePossiblyWrappedJson('not-json')).toBeNull();
	});

	it('formats banned messages with fallbacks', () => {
		expect(
			formatBannedMessage({ reason: ' spam ', until: ' t ' }, (r, u) => `${r}|${u}`),
		).toBe('spam|t');
		expect(formatBannedMessage({ error: 'banned' }, (r, u) => `${r}:${u}`)).toBe('banned:');
	});
});
