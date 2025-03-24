export const formatDateTime = (dateStr: string): string => {
	const date = new Date(dateStr);
	return `${addZero(date.getHours())}:${addZero(date.getMinutes())}:${addZero(date.getSeconds())} ${addZero(date.getDate())}-${addZero(date.getMonth())}-${date.getFullYear()}`

}

const addZero = (num: number): string => {
	return num < 10 ? "0" + num : "" + num
}