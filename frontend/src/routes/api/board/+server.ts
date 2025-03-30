import { MAIN_BACKEND_URL } from "$env/static/private";
import type { Board } from "$lib/types/board";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({request, fetch}) => {
	const body: Board = await request.json();

	const res = await fetch(`${MAIN_BACKEND_URL}/api/v1/board`, {
		method: 'POST',
		body: JSON.stringify(body),
	})

	return res;
}; 