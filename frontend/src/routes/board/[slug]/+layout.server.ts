import { MAIN_BACKEND_URL } from '$env/static/private';
import { trimLargeWords } from '$lib/helpers';
import type { Board } from '$lib/types/board';
import type { LayoutServerLoad } from './$types';

const DEFAULT_LIMIT = 20;

export const load: LayoutServerLoad = async ({ params, url, depends, fetch, cookies }) => {
	const { slug } = params;

	depends('board:slug');

	const fingerprint = cookies.get('fingerprint');

	const limitParam = Number(url.searchParams.get('limit') ?? DEFAULT_LIMIT);
	const offsetParam = Number(url.searchParams.get('offset') ?? 0);
	const limit = Number.isFinite(limitParam) && limitParam > 0 ? Math.min(limitParam, 100) : DEFAULT_LIMIT;
	const offset = Number.isFinite(offsetParam) && offsetParam >= 0 ? offsetParam : 0;

	const res = await fetch(
		`${MAIN_BACKEND_URL}/api/v1/board/${slug}?limit=${limit}&offset=${offset}`,
	);
	const board: Board = await res.json();
	board.threads.forEach((thread) => {
		if (thread.original_post.browser_fingerprint === fingerprint) {
			thread.is_author = true;
		}
		if (thread.original_post.text.length > 50) {
			thread.original_post.text = trimLargeWords(thread.original_post.text);
		}
		thread.original_post.browser_fingerprint = '';
	});

	return { slug, board };
};
