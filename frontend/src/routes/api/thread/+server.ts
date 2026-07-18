import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import {
	adminBackendHeaders,
	assertBodySize,
	assertCaptcha,
	assertMultipartFiles,
	assertText,
	assertTitle,
	fingerprintFromCookies,
	proxyBackend,
} from '$lib/server/backend';

export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	assertBodySize(request);

	const form = await request.formData();
	const captcha = {
		token: String(form.get('captcha_token') ?? ''),
		input: String(form.get('captcha_input') ?? ''),
	};
	assertCaptcha(captcha);
	assertTitle(String(form.get('title') ?? ''));
	assertText(String(form.get('text') ?? ''));

	const boardId = String(form.get('board_id') ?? '');
	if (!boardId) {
		error(400, { message: 'board_id is required' });
	}

	const files = form.getAll('files').filter((f): f is File => f instanceof File && f.size > 0);
	assertMultipartFiles(files);

	// Captcha is consumed once on the Go API — do not pre-validate here.
	const out = new FormData();
	out.set('board_id', boardId);
	out.set('title', String(form.get('title') ?? ''));
	out.set('text', String(form.get('text') ?? ''));
	out.set('browser_fingerprint', fingerprintFromCookies(cookies));
	out.set('captcha_token', captcha.token);
	out.set('captcha_input', captcha.input);
	for (const file of files) {
		out.append('files', file, file.name);
	}

	return proxyBackend(fetch, '/api/v1/thread', {
		method: 'POST',
		headers: await adminBackendHeaders(cookies),
		body: out,
	});
};
