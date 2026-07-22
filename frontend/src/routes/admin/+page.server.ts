import { env } from '$env/dynamic/private';
import { adminBackendHeaders } from '$lib/server/backend';
import type { Cookies } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

const MAIN_BACKEND_URL = env.MAIN_BACKEND_URL;

interface BoardMetrics {
	alias: string;
	post_count: number;
	deleted_count: number;
	sage_count: number;
	thread_count: number;
}

interface DailyMetrics {
	date: string;
	boards: BoardMetrics[];
}

async function fetchDailyMetrics(fetch: typeof globalThis.fetch, cookies: Cookies, date: Date): Promise<BoardMetrics[]> {
	const from = new Date(date);
	from.setHours(0, 0, 0, 0);
	const to = new Date(date);
	to.setHours(23, 59, 59, 999);

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/metric/posts?from=${from.toISOString()}&to=${to.toISOString()}`, {
		headers: await adminBackendHeaders(cookies),
	});

	if (!res.ok) {
		return [];
	}

	const data = await res.json();
	return data.boards || [];
}

export const load: PageServerLoad = async ({ fetch, cookies, depends }) => {
	depends('admin:metrics');

	const now = new Date();
	const dailyMetrics: DailyMetrics[] = [];

	for (let i = 6; i >= 0; i--) {
		const date = new Date(now);
		date.setDate(date.getDate() - i);
		date.setHours(0, 0, 0, 0);

		const boards = await fetchDailyMetrics(fetch, cookies, date);
		dailyMetrics.push({
			date: date.toISOString().split('T')[0],
			boards
		});
	}

	return { dailyMetrics, error: null };
};