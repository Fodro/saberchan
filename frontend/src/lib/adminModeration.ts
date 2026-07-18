import { readApiError } from '$lib/composeValidate';

export async function softDelete(kind: 'board' | 'thread' | 'post', id: string): Promise<string | null> {
	const res = await fetch(`/api/admin/${kind}/${id}`, { method: 'DELETE' });
	if (res.ok || res.status === 204 || res.status === 200) return null;
	return readApiError(res);
}

export async function restoreDeleted(
	kind: 'board' | 'thread' | 'post',
	id: string,
): Promise<string | null> {
	const res = await fetch(`/api/admin/${kind}/${id}/restore`, { method: 'POST' });
	if (res.ok || res.status === 204 || res.status === 200) return null;
	return readApiError(res);
}

export type BanDuration = '1h' | '1d' | '7d' | '30d' | 'permanent';

export async function banPost(
	id: string,
	reason: string,
	duration: BanDuration,
): Promise<string | null> {
	const res = await fetch(`/api/admin/post/${id}/ban`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ reason, duration }),
	});
	if (res.ok || res.status === 204 || res.status === 200) return null;
	return readApiError(res);
}
