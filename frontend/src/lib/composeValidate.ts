import { bufferToBase64 } from '$lib/helpers';
import {
	MAX_FILE_BYTES,
	MAX_FILES,
	MAX_JSON_BODY_CHARS,
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
	| 'payload_too_large';

export type AttachmentPayload = {
	name: string;
	type: string;
	body: string;
};

export function buildAttachments(files: FileType[]): AttachmentPayload[] {
	return files.map((value) => ({
		name: value.name,
		type: 'image',
		body: bufferToBase64(value.blob),
	}));
}

export function validateCompose(input: {
	title?: string | null;
	text?: string | null;
	requireTitle?: boolean;
	captchaInput?: string;
	captchaToken?: string;
	files: FileType[];
	attachments: AttachmentPayload[];
	payload: unknown;
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
		const size =
			typeof file.blob === 'string'
				? file.blob.length
				: file.blob instanceof ArrayBuffer
					? file.blob.byteLength
					: 0;
		if (size > MAX_FILE_BYTES) return 'file_size';
	}

	const encoded = JSON.stringify(input.payload);
	if (encoded.length > MAX_JSON_BODY_CHARS) return 'payload_too_large';

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
