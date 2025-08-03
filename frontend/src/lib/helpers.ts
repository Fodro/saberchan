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

export const bufferToBase64 = (buf: string | ArrayBuffer | null): string => {
	if (!buf) {
		return "";
	}
	if (typeof buf === 'string') {
		return buf
	}
	const chunks = [];
	const uint8 = new Uint8Array(buf);
	const chunkSize = 0x8000;
	for (let i = 0; i < uint8.length; i += chunkSize) {
		const chunk = uint8.subarray(i, Math.min(i + chunkSize, uint8.length));
		chunks.push(String.fromCharCode(...chunk));
	}
	return btoa(chunks.join(''));
}

export const  base64ToArrayBuffer = (base64: string)  => {
	const binaryString = atob(base64);
	const bytes = new Uint8Array(binaryString.length);
	for (let i = 0; i < binaryString.length; i++) {
		bytes[i] = binaryString.charCodeAt(i);
	}
	return bytes.buffer;
}

const insertStringAtIndex = (originalString: string, stringToInsert: string, index: number): string => {
	const firstPart = originalString.slice(0, index);

	const secondPart = originalString.slice(index);

	return firstPart + stringToInsert + secondPart;
}


export const trimLargeWords = (str: string): string => {
	const words = str.split(' ');
	const trimmedWords = words.map(word => {
		if (word.length > 50) {
			for (let i = 24; i < word.length; i+=25) {
				word = insertStringAtIndex(word, '\n', i);
			}
		}
		return word;
	});
	return trimmedWords.join(' ');
}