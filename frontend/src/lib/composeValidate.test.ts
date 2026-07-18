import { describe, expect, it } from 'vitest';
import {
	MAX_FILES,
	MAX_IMAGE_BYTES,
	MAX_TEXT_CHARS,
	MAX_TITLE_CHARS,
} from './limits';
import { mimeFromName, validateCompose } from './composeValidate';
import type { File as FileType } from './types/attachment';

function file(name: string, size: number): FileType {
	return { id: name, name, blob: new ArrayBuffer(size) };
}

const okBase = {
	text: 'hello',
	captchaInput: '1',
	captchaToken: 'tok',
	files: [] as FileType[],
};

describe('mimeFromName', () => {
	it('maps known extensions', () => {
		expect(mimeFromName('a.jpg')).toBe('image/jpeg');
		expect(mimeFromName('a.JPEG')).toBe('image/jpeg');
		expect(mimeFromName('a.png')).toBe('image/png');
		expect(mimeFromName('a.webm')).toBe('video/webm');
		expect(mimeFromName('a.mp4')).toBe('video/mp4');
	});

	it('falls back for unknown', () => {
		expect(mimeFromName('a.exe')).toBe('application/octet-stream');
		expect(mimeFromName('noext')).toBe('application/octet-stream');
	});
});

describe('validateCompose', () => {
	it('requires text and captcha', () => {
		expect(validateCompose({ ...okBase, text: '  ' })).toBe('text_required');
		expect(validateCompose({ ...okBase, captchaInput: '' })).toBe('captcha_required');
		expect(validateCompose({ ...okBase, captchaToken: ' ' })).toBe('captcha_required');
	});

	it('validates title when required', () => {
		expect(validateCompose({ ...okBase, requireTitle: true, title: '' })).toBe('title_required');
		expect(
			validateCompose({
				...okBase,
				requireTitle: true,
				title: 'x'.repeat(MAX_TITLE_CHARS + 1),
			}),
		).toBe('title_too_long');
		expect(validateCompose({ ...okBase, requireTitle: true, title: 'ok' })).toBeNull();
	});

	it('rejects oversized text', () => {
		expect(validateCompose({ ...okBase, text: 'x'.repeat(MAX_TEXT_CHARS + 1) })).toBe(
			'text_too_long',
		);
	});

	it('rejects too many / bad / oversized files', () => {
		expect(
			validateCompose({
				...okBase,
				files: Array.from({ length: MAX_FILES + 1 }, (_, i) => file(`a${i}.jpg`, 10)),
			}),
		).toBe('file_count');
		expect(validateCompose({ ...okBase, files: [file('x.exe', 10)] })).toBe('file_type');
		expect(
			validateCompose({ ...okBase, files: [file('big.jpg', MAX_IMAGE_BYTES + 1)] }),
		).toBe('file_size');
		expect(validateCompose({ ...okBase, files: [file('ok.webm', 100)] })).toBeNull();
	});

	it('accepts a valid reply payload', () => {
		expect(validateCompose(okBase)).toBeNull();
	});
});
