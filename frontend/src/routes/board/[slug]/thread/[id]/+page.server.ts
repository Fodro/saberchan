import { MAIN_BACKEND_URL } from '$env/static/private';
import type { Thread } from '$lib/types/thread';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, depends, fetch }) => {
	const { slug, id } = params;

	depends("thread:id");

	const resThread = await fetch(`${MAIN_BACKEND_URL}/api/v1/thread/${id}`);
	const thread: Thread = await resThread.json();

	return { slug, thread };
};