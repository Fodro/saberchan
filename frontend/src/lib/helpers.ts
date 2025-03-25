export const formatDateTime = (dateStr: string): string => {
	const date = new Date(dateStr);
	return `${addZero(date.getHours())}:${addZero(date.getMinutes())}:${addZero(date.getSeconds())} ${addZero(date.getDate())}-${addZero(date.getMonth())}-${date.getFullYear()}`

}

const addZero = (num: number): string => {
	return num < 10 ? "0" + num : "" + num
}

export const scrollIntoView = (id: string) => {
	const el = document.getElementById(id);
	if (!el) return;
	el.scrollIntoView({
		behavior: 'smooth'
	});
}

export const insertTagAtCursor = (myField: HTMLTextAreaElement, open: string, close: string) => {
	if (myField.selectionStart || myField.selectionStart === 0) {
		const startPos = myField.selectionStart;
		const endPos = myField.selectionEnd;
		myField.value = myField.value.substring(0, startPos) + open + myField.value.substring(startPos, endPos) + close + myField.value.substring(endPos, myField.value.length);
	} else {
		myField.value += open + close;
	}
}