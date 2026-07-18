export const trackBoard = async (alias: string) => {
	return await fetch("/tracking/board", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({
			alias: alias,
		}),
	});
};