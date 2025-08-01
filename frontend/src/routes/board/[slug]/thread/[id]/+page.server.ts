import { MAIN_BACKEND_URL } from '$env/static/private';
import { trimLargeWords } from '$lib/helpers';
import type { Thread } from '$lib/types/thread';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, depends, fetch, cookies }) => {
	const { slug, id } = params;

	depends("thread:id");

	const fingerprint = cookies.get('fingerprint');

	const resThread = await fetch(`${MAIN_BACKEND_URL}/api/v1/thread/${id}`);
	let thread: Thread;
	try {
	 thread = await resThread.json();
	} catch (error) {
		console.error({error, id});
		return { slug }
	}

	if (thread.original_post.browser_fingerprint === fingerprint) {
		thread.original_post.is_author = true;
	}

	thread.original_post.browser_fingerprint = ""

	if (thread.original_post.text.length > 50) {
		thread.original_post.text = trimLargeWords(thread.original_post.text)
	}

	thread.posts.forEach((post) => {
		if (post.browser_fingerprint === fingerprint) {
			post.is_author = true;
		}
		if (post.text.length > 50) {
			post.text = trimLargeWords(post.text)
		}
		post.browser_fingerprint = ""
	})

	return { slug, thread };
};