/** Shared upload / compose limits (keep in sync with FileUploader + BFF + BE). */
export const MAX_IMAGE_BYTES = 5 * 1024 * 1024; // 5 MiB
export const MAX_VIDEO_BYTES = 10 * 1024 * 1024; // 10 MiB
/** @deprecated Prefer MAX_IMAGE_BYTES / maxBytesForMime. */
export const MAX_FILE_BYTES = MAX_IMAGE_BYTES;
export const MAX_FILES = 4;
export const MAX_TITLE_CHARS = 255;
export const MAX_TEXT_CHARS = 16_000;
/** Overall request body cap under adapter BODY_SIZE_LIMIT (16M+). */
export const MAX_JSON_BODY_CHARS = 12 * 1024 * 1024;

export const IMAGE_MIME = new Set([
	'image/jpeg',
	'image/jpg',
	'image/png',
	'image/gif',
	'image/webp',
]);

export const VIDEO_MIME = new Set(['video/webm', 'video/mp4']);

export const VIDEO_EXT = new Set(['webm', 'mp4']);

export function maxBytesForMime(mime: string): number {
	if (VIDEO_MIME.has(mime)) return MAX_VIDEO_BYTES;
	return MAX_IMAGE_BYTES;
}

export function isVideoMime(mime: string | undefined | null): boolean {
	if (!mime) return false;
	return VIDEO_MIME.has(mime);
}

/** Prefer stored MIME; fall back to filename extension for older rows. */
export function isVideoAttachment(name: string, type?: string | null): boolean {
	if (type && VIDEO_MIME.has(type)) return true;
	const ext = name.split('.').pop()?.toLowerCase() ?? '';
	return VIDEO_EXT.has(ext);
}
