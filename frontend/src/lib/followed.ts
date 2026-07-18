import { invalidate } from '$app/navigation';

async function readError(res: Response): Promise<string> {
	try {
		const body = (await res.json()) as { message?: string };
		if (body?.message) return body.message;
	} catch {
		/* ignore */
	}
	return `Request failed (${res.status})`;
}

/** Follow a thread: write cookie via BFF and POST backend Redis key. */
export async function followThread(id: string, lastSeenReplies?: number): Promise<void> {
	const res = await fetch(`/api/follow/${encodeURIComponent(id)}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(
			lastSeenReplies !== undefined ? { lastSeenReplies } : {},
		),
	});
	if (!res.ok) {
		throw new Error(await readError(res));
	}
	await invalidate('followed:list');
}

/** Unfollow: remove from cookie only (does not clear Redis). */
export async function unfollowThread(id: string): Promise<void> {
	const res = await fetch(`/api/follow/${encodeURIComponent(id)}`, {
		method: 'DELETE',
	});
	if (!res.ok) {
		throw new Error(await readError(res));
	}
	await invalidate('followed:list');
}
