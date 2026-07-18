import {
	fileBlob,
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
			case 'file_type':
				return t('common.file.limitType');
		}
	};
}

export type SubmitComposeResult =
	| { ok: true; status: number; json: unknown }
	| {
			ok: false;
			kind: 'captcha' | 'validation' | 'http' | 'banned';
			message: string;
			bumpCaptcha?: boolean;
	  };

type BanPayload = { code?: string; reason?: string; until?: string; error?: string; message?: string };

function parsePossiblyWrappedJson(raw: string): BanPayload | null {
	try {
		const outer = JSON.parse(raw) as BanPayload;
		if (typeof outer.message === 'string' && outer.message.trim().startsWith('{')) {
			try {
				return JSON.parse(outer.message) as BanPayload;
			} catch {
				return outer;
			}
		}
		return outer;
	} catch {
		return null;
	}
}

function formatBannedMessage(
	payload: BanPayload,
	bannedMessage: (reason: string, until: string) => string,
): string {
	const reason = payload.reason?.trim() || payload.error || 'banned';
	const until = payload.until?.trim() || '';
	return bannedMessage(reason, until);
}

export async function submitCompose(opts: {
	endpoint: '/api/thread' | '/api/post';
	fields: Record<string, string>;
	title?: string | null;
	text?: string | null;
	requireTitle?: boolean;
	captchaInput: string;
	captchaToken: string;
	files: FileType[];
	errorMessage: ComposeErrorMessageFn;
	captchaFailedMessage: string;
	bannedMessage?: (reason: string, until: string) => string;
}): Promise<SubmitComposeResult> {
	const invalid = validateCompose({
		title: opts.title,
		text: opts.text,
		requireTitle: opts.requireTitle,
		captchaInput: opts.captchaInput,
		captchaToken: opts.captchaToken,
		files: opts.files,
	});
	if (invalid) {
		return { ok: false, kind: 'validation', message: opts.errorMessage(invalid) };
	}

	const form = new FormData();
	for (const [key, value] of Object.entries(opts.fields)) {
		form.set(key, value);
	}
	form.set('captcha_input', opts.captchaInput);
	form.set('captcha_token', opts.captchaToken);
	for (const file of opts.files) {
		form.append('files', fileBlob(file), file.name);
	}

	const res = await fetch(opts.endpoint, {
		method: 'POST',
		body: form,
	});

	if (res.status === 403) {
		const raw = await res.text();
		const payload = parsePossiblyWrappedJson(raw);
		if (payload?.code === 'banned') {
			const bannedMessage =
				opts.bannedMessage ??
				((reason, until) => (until ? `${reason} (until ${until})` : reason));
			return {
				ok: false,
				kind: 'banned',
				message: formatBannedMessage(payload, bannedMessage),
			};
		}
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

export { fileBlob, readApiError, validateCompose };
export type { ComposeValidationError };
