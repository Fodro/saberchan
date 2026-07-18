import { MAIN_BACKEND_URL, ADMIN_API_TOKEN } from '$env/static/private';
import { error } from '@sveltejs/kit';
import {
	IMAGE_MIME,
	MAX_FILES,
	MAX_JSON_BODY_CHARS,
	MAX_TEXT_CHARS,
	MAX_TITLE_CHARS,
	VIDEO_MIME,
	maxBytesForMime,
} from '$lib/limits';
import { verifyAccessToken } from '$lib/server/oidc';
import type { Cookies } from '@sveltejs/kit';

export function backendUrl(path: string): string {
	const base = MAIN_BACKEND_URL.replace(/\/$/, '');
	const p = path.startsWith('/') ? path : `/${path}`;
	return `${base}${p}`;
}

export function assertBodySize(request: Request) {
	const contentLength = Number(request.headers.get('content-length') || 0);
	if (contentLength > MAX_JSON_BODY_CHARS) {
		error(413, { message: 'Post is too large (try fewer or smaller images)' });
	}
}

export function assertCaptcha(captcha: { token?: string; input?: string } | undefined) {
	if (!captcha?.token?.trim() || !captcha?.input?.trim()) {
		error(400, { message: 'Solve the captcha before posting' });
	}
}

export function assertText(text: string | null | undefined) {
	const t = text?.trim() ?? '';
	if (!t) error(400, { message: 'Post text is required' });
	if (t.length > MAX_TEXT_CHARS) {
		error(400, { message: `Post text is too long (max ${MAX_TEXT_CHARS} characters)` });
	}
}

export function assertTitle(title: string | null | undefined) {
	const t = title?.trim() ?? '';
	if (!t) error(400, { message: 'Title is required' });
	if (t.length > MAX_TITLE_CHARS) {
		error(400, { message: `Title is too long (max ${MAX_TITLE_CHARS} characters)` });
	}
}

export function assertMultipartFiles(files: File[]) {
	if (files.length > MAX_FILES) {
		error(400, { message: `Maximum file count is ${MAX_FILES}` });
	}
	for (const file of files) {
		const mime = file.type || '';
		const allowed = IMAGE_MIME.has(mime) || VIDEO_MIME.has(mime);
		if (mime && !allowed) {
			error(400, { message: `Unsupported file type: ${file.type}` });
		}
		const limit = maxBytesForMime(mime || 'image/jpeg');
		if (file.size > limit) {
			const label = VIDEO_MIME.has(mime) ? '10MB for videos' : '5MB for images';
			error(413, { message: `Maximum file size is ${label}` });
		}
	}
}

export function fingerprintFromCookies(cookies: Cookies): string {
	return cookies.get('fingerprint') || '';
}

/** True when the accessToken cookie is a valid Keycloak JWT for our client. */
export async function isAdminSession(cookies: Cookies): Promise<boolean> {
	const token = cookies.get('accessToken');
	if (!token) return false;
	return verifyAccessToken(token);
}

/** Headers for Go admin-gated routes (create board, locked-board thread, etc.). */
export async function adminBackendHeaders(cookies: Cookies): Promise<Record<string, string>> {
	if (!(await isAdminSession(cookies)) || !ADMIN_API_TOKEN) return {};
	return { 'X-Admin-Token': ADMIN_API_TOKEN };
}

export async function proxyBackend(
	fetchFn: typeof fetch,
	path: string,
	init?: RequestInit,
): Promise<Response> {
	const res = await fetchFn(backendUrl(path), init);
	if (!res.ok) {
		const msg = await res.text();
		error(res.status, { message: msg || `Backend request failed (${res.status})` });
	}
	return res;
}
