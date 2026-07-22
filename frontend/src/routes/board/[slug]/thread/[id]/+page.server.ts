import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { trimLargeWords } from '$lib/helpers';
import { adminBackendHeaders } from '$lib/server/backend';

const MAIN_BACKEND_URL = env.MAIN_BACKEND_URL;
import {
	FOLLOWED_COOKIE,
	markSeen,
	parseFollowedCookie,
	setFollowedCookie,
} from '$lib/server/followed';
import type { Thread } from '$lib/types/thread';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, depends, fetch, cookies }) => {
	const { slug, id } = params;

	depends('thread:id');

	const fingerprint = cookies.get('fingerprint');

	const resThread = await fetch(`${MAIN_BACKEND_URL}/api/v1/thread/${id}`, {
		headers: await adminBackendHeaders(cookies),
	});
	if (!resThread.ok) {
		redirect(302, '/404');
	}

	let thread: Thread;
	try {
		thread = await resThread.json();
	} catch (error) {
		console.error({ error, id });
		redirect(302, '/404');
	}

	if (thread.original_post.browser_fingerprint === fingerprint) {
		thread.original_post.is_author = true;
	}

	thread.original_post.browser_fingerprint = '';

	if (thread.original_post.text.length > 50) {
		thread.original_post.text = trimLargeWords(thread.original_post.text);
	}

	thread.posts.forEach((post) => {
		if (post.browser_fingerprint === fingerprint) {
			post.is_author = true;
		}
		if (post.text.length > 50) {
			post.text = trimLargeWords(post.text);
		}
		post.browser_fingerprint = '';
	});

	// Mark followed thread as seen at current reply count (posts = replies, OP separate).
	const replies = thread.posts?.length ?? thread.replies_count ?? 0;
	const updated = markSeen(parseFollowedCookie(cookies.get(FOLLOWED_COOKIE)), id, replies);
	if (updated) {
		setFollowedCookie(cookies, updated);
	}

	return { slug, thread };
};
