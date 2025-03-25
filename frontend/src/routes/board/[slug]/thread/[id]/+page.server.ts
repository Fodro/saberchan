import { MAIN_BACKEND_URL } from '$env/static/private';
import type { Thread } from '$lib/types/thread';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, depends, fetch, cookies }) => {
	const { slug, id } = params;

	depends("thread:id");

	const fingerprint = cookies.get('fingerprint');

	const resThread = await fetch(`${MAIN_BACKEND_URL}/api/v1/thread/${id}`);
	const thread: Thread = await resThread.json();

	if (thread.original_post.browser_fingerprint === fingerprint) {
		thread.original_post.is_author = true;
	}

	thread.original_post.browser_fingerprint = ""

	thread.posts.forEach((post) => {
		if (post.browser_fingerprint === fingerprint) {
			post.is_author = true;
		}
		post.browser_fingerprint = ""
	})

	return { slug, thread };
};