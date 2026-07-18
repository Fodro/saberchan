import { describe, expect, it } from 'vitest';
import {
	IMAGE_MIME,
	MAX_IMAGE_BYTES,
	MAX_VIDEO_BYTES,
	VIDEO_MIME,
	isVideoAttachment,
	isVideoMime,
	maxBytesForMime,
} from './limits';

describe('maxBytesForMime', () => {
	it('returns video cap for video MIME', () => {
		expect(maxBytesForMime('video/webm')).toBe(MAX_VIDEO_BYTES);
		expect(maxBytesForMime('video/mp4')).toBe(MAX_VIDEO_BYTES);
	});

	it('returns image cap otherwise', () => {
		expect(maxBytesForMime('image/jpeg')).toBe(MAX_IMAGE_BYTES);
		expect(maxBytesForMime('application/octet-stream')).toBe(MAX_IMAGE_BYTES);
	});
});

describe('isVideoMime / isVideoAttachment', () => {
	it('detects video MIME', () => {
		expect(isVideoMime('video/mp4')).toBe(true);
		expect(isVideoMime('image/png')).toBe(false);
		expect(isVideoMime(null)).toBe(false);
		expect(isVideoMime(undefined)).toBe(false);
	});

	it('falls back to extension when type missing', () => {
		expect(isVideoAttachment('clip.webm')).toBe(true);
		expect(isVideoAttachment('clip.mp4', '')).toBe(true);
		expect(isVideoAttachment('pic.png')).toBe(false);
		expect(isVideoAttachment('x.bin', 'video/webm')).toBe(true);
	});

	it('keeps mime sets populated', () => {
		expect(IMAGE_MIME.has('image/png')).toBe(true);
		expect(VIDEO_MIME.has('video/webm')).toBe(true);
	});
});
