import {
	buildAttachments,
	readApiError,
	validateCompose,
	type ComposeValidationError,
} from '$lib/composeValidate';
import type { File as FileType } from '$lib/types/attachment';

export type ComposeErrorMessageFn = (code: ComposeValidationError) => string;

export function composeErrorMessageFactory(t: (key: string) => string): ComposeErrorMessageFn {
	return (code) => {
		switch (code) {
			case 'title_required':
				return t('common.compose.title_required');
			case 'title_too_long':
				return t('common.compose.title_too_long');
			case 'text_required':
				return t('common.compose.text_required');
			case 'text_too_long':
				return t('common.compose.text_too_long');
			case 'captcha_required':
				return t('common.captcha.required');
			case 'file_count':
				return t('common.file.limitCount');
			case 'file_size':
				return t('common.file.limitSize');
			case 'payload_too_large':
				return t('common.compose.payload_too_large');
		}
	};
}

export type SubmitComposeResult =
	| { ok: true; status: number; json: unknown }
	| { ok: false; kind: 'captcha' | 'validation' | 'http'; message: string; bumpCaptcha?: boolean };

export async function submitCompose(opts: {
	endpoint: '/api/thread' | '/api/post';
	payload: unknown;
	title?: string | null;
	text?: string | null;
	requireTitle?: boolean;
	captchaInput: string;
	captchaToken: string;
	files: FileType[];
	errorMessage: ComposeErrorMessageFn;
	captchaFailedMessage: string;
}): Promise<SubmitComposeResult> {
	const attachments = buildAttachments(opts.files);
	const invalid = validateCompose({
		title: opts.title,
		text: opts.text,
		requireTitle: opts.requireTitle,
		captchaInput: opts.captchaInput,
		captchaToken: opts.captchaToken,
		files: opts.files,
		attachments,
		payload: opts.payload,
	});
	if (invalid) {
		return { ok: false, kind: 'validation', message: opts.errorMessage(invalid) };
	}

	const res = await fetch(opts.endpoint, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(opts.payload),
	});

	if (res.status === 403) {
		return {
			ok: false,
			kind: 'captcha',
			message: opts.captchaFailedMessage,
			bumpCaptcha: true,
		};
	}

	if (res.status !== 201 && res.status !== 200) {
		return { ok: false, kind: 'http', message: await readApiError(res) };
	}

	const json = await res.json().catch(() => null);
	return { ok: true, status: res.status, json };
}

export { buildAttachments, readApiError, validateCompose };
export type { ComposeValidationError };
