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

export const insertTagAtCursor = (field: HTMLTextAreaElement, open: string, close: string) => {
	if (field.selectionStart || field.selectionStart === 0) {
		const startPos = field.selectionStart;
		const endPos = field.selectionEnd;
		field.value = field.value.substring(0, startPos) + open + field.value.substring(startPos, endPos) + close + field.value.substring(endPos, field.value.length);
	} else {
		field.value += open + close;
	}
	field.dispatchEvent(new Event('input', {}));
}

export const verifyExp = (exp: number | undefined): boolean => {
	if (!exp) {
		return true;
	}

	return Date.now() > exp * 1000;
}