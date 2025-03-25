import type { Locale } from "$lib/types/metadata";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({ request, cookies }) => {
	const body: Locale = await request.json();
	const { locale } = body;
	cookies.set("locale", locale, { path: '/' });

	return new Response('', {
		status: 200
	});
}; 