export const trackBoard = async (alias: string) => {
	return await fetch("/tracking/board", {
		method: "POST",
		body: JSON.stringify({
			alias: alias,
		}),
	});
}