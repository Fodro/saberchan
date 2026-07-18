import { describe, expect, it } from 'vitest';
import {
	buildPostLines,
	classifyLine,
	parseMarkup,
	replyHref,
	type MarkupNode,
	type MarkupTag,
} from './markup';

function text(value: string): MarkupNode {
	return { type: 'text', value };
}

function tag(name: MarkupTag, children: MarkupNode[]): MarkupNode {
	return { type: 'tag', tag: name, children };
}

describe('parseMarkup', () => {
	it('returns plain text when there are no tags', () => {
		expect(parseMarkup('hello world')).toEqual([text('hello world')]);
	});

	it('parses a single tag', () => {
		expect(parseMarkup('[b]bold[/b]')).toEqual([tag('b', [text('bold')])]);
	});

	it('parses nested tags', () => {
		expect(parseMarkup('[b]a [i]b[/i] c[/b]')).toEqual([
			tag('b', [text('a '), tag('i', [text('b')]), text(' c')]),
		]);
	});

	it('parses adjacent tags and surrounding text', () => {
		expect(parseMarkup('x[b]y[/b]z')).toEqual([text('x'), tag('b', [text('y')]), text('z')]);
	});

	it('leaves unknown tags as literal text', () => {
		expect(parseMarkup('[foo]bar[/foo]')).toEqual([text('[foo]bar[/foo]')]);
	});

	it('leaves unclosed tags content without wrapping', () => {
		// Opener consumed, children parsed until EOF (no close) — still a tag node.
		expect(parseMarkup('[b]oops')).toEqual([tag('b', [text('oops')])]);
	});

	it('treats mismatched closers as literal text inside the open tag', () => {
		expect(parseMarkup('[b]x[/i]y[/b]')).toEqual([tag('b', [text('x[/i]y')])]);
	});

	it('handles spoiler / sup / sub', () => {
		expect(parseMarkup('[spoiler]hide[/spoiler][sup]1[/sup][sub]2[/sub]')).toEqual([
			tag('spoiler', [text('hide')]),
			tag('sup', [text('1')]),
			tag('sub', [text('2')]),
		]);
	});

	it('does not interpret raw HTML as markup', () => {
		expect(parseMarkup('<script>alert(1)</script>')).toEqual([
			text('<script>alert(1)</script>'),
		]);
	});

	it('handles incomplete bracket as literal', () => {
		expect(parseMarkup('a[b')).toEqual([text('a[b')]);
	});
});

describe('classifyLine / replyHref', () => {
	it('classifies reply, greentext, and plain', () => {
		expect(classifyLine('>>12')).toBe('reply');
		expect(classifyLine('>greentext')).toBe('greentext');
		expect(classifyLine('hello')).toBe('plain');
		expect(classifyLine('')).toBe('plain');
	});

	it('builds reply href like the old PostBody', () => {
		expect(replyHref('>>123')).toBe('#123');
		expect(replyHref('>>1 >>2')).toBe('#1 2');
	});
});

describe('buildPostLines', () => {
	it('splits on newlines and trims', () => {
		const lines = buildPostLines('  [b]a[/b]  \n>hi\n>>9');
		expect(lines).toHaveLength(3);
		expect(lines[0]).toMatchObject({
			trimmed: '[b]a[/b]',
			kind: 'plain',
			nodes: [tag('b', [text('a')])],
		});
		expect(lines[1]).toMatchObject({
			trimmed: '>hi',
			kind: 'greentext',
		});
		expect(lines[2]).toMatchObject({
			trimmed: '>>9',
			kind: 'reply',
			href: '#9',
		});
	});

	it('preserves empty lines', () => {
		expect(buildPostLines('a\n\nb').map((l) => l.trimmed)).toEqual(['a', '', 'b']);
	});
});
