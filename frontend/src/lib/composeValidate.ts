import {
	MAX_FILE_BYTES,
	MAX_FILES,
	MAX_TEXT_CHARS,
	MAX_TITLE_CHARS,
} from '$lib/limits';
import type { File as FileType } from '$lib/types/attachment';

export type ComposeValidationError =
	| 'title_required'
	| 'title_too_long'
	| 'text_required'
	| 'text_too_long'
	| 'captcha_required'
	| 'file_count'
	| 'file_size'
	| 'file_type';

const ALLOWED_EXT = new Set(['jpg', 'jpeg', 'png', 'gif', 'webp']);

export function mimeFromName(name: string): string {
	const ext = name.split('.').pop()?.toLowerCase() ?? '';
	switch (ext) {
		case 'jpg':
		case 'jpeg':
			return 'image/jpeg';
		case 'png':
			return 'image/png';
		case 'gif':
			return 'image/gif';
		case 'webp':
			return 'image/webp';
		default:
			return 'application/octet-stream';
	}
}

export function fileBlob(file: FileType): Blob {
	const type = mimeFromName(file.name);
	if (file.blob instanceof ArrayBuffer) {
		return new Blob([file.blob], { type });
	}
	if (typeof file.blob === 'string') {
		// Legacy: treat as binary string length for size checks only — prefer ArrayBuffer.
		const buf = Uint8Array.from(file.blob, (c) => c.charCodeAt(0));
		return new Blob([buf], { type });
	}
	return new Blob([], { type });
}

export function validateCompose(input: {
	title?: string | null;
	text?: string | null;
	requireTitle?: boolean;
	captchaInput?: string;
	captchaToken?: string;
	files: FileType[];
}): ComposeValidationError | null {
	const title = input.title?.trim() ?? '';
	const text = input.text?.trim() ?? '';

	if (input.requireTitle) {
		if (!title) return 'title_required';
		if (title.length > MAX_TITLE_CHARS) return 'title_too_long';
	}

	if (!text) return 'text_required';
	if (text.length > MAX_TEXT_CHARS) return 'text_too_long';

	if (!input.captchaToken?.trim() || !input.captchaInput?.trim()) {
		return 'captcha_required';
	}

	if (input.files.length > MAX_FILES) return 'file_count';
	for (const file of input.files) {
		const ext = file.name.split('.').pop()?.toLowerCase() ?? '';
		if (!ALLOWED_EXT.has(ext)) return 'file_type';
		const size =
			typeof file.blob === 'string'
				? file.blob.length
				: file.blob instanceof ArrayBuffer
					? file.blob.byteLength
					: 0;
		if (size > MAX_FILE_BYTES) return 'file_size';
	}

	return null;
}

export async function readApiError(res: Response): Promise<string> {
	const raw = await res.text();
	try {
		const json = JSON.parse(raw) as { message?: string; error?: string };
		return json.message || json.error || raw || `Request failed (${res.status})`;
	} catch {
		return raw || `Request failed (${res.status})`;
	}
}
