import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ url }) => {
	const imgLink = url.searchParams.get("link") || "";
	const imgName = url.searchParams.get("name") || "";

	return ({
		imgLink,
		imgName
	});
};