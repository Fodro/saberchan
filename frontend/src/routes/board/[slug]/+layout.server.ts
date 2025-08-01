import { MAIN_BACKEND_URL } from '$env/static/private';
import { trimLargeWords } from '$lib/helpers';
import type { Board } from '$lib/types/board';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ params, depends, fetch, cookies }) => {
	const { slug } = params;

	depends("board:slug");

	const fingerprint = cookies.get('fingerprint');

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/board/${slug}`);
	const board: Board = await res.json();
	board.threads.forEach((thread) => {
		if (thread.original_post.browser_fingerprint === fingerprint) {
			thread.is_author = true;
		}
		if (thread.original_post.text.length > 50) {
			thread.original_post.text = trimLargeWords(thread.original_post.text)
		}
		thread.original_post.browser_fingerprint = ""
	})

	return { slug, board };
};