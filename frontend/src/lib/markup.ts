/** BBCode tags supported by the compose toolbar / PostBody. */
export const MARKUP_TAGS = ['b', 'i', 'u', 'o', 's', 'sup', 'sub', 'spoiler'] as const;

export type MarkupTag = (typeof MARKUP_TAGS)[number];

export type MarkupNode =
	| { type: 'text'; value: string }
	| { type: 'tag'; tag: MarkupTag; children: MarkupNode[] };

const TAG_SET = new Set<string>(MARKUP_TAGS);

/** Tailwind classes previously applied via sanitize-html spans. */
export const MARKUP_TAG_CLASS: Record<MarkupTag, string> = {
	b: 'font-bold',
	i: 'italic',
	u: 'underline',
	o: 'overline',
	s: 'line-through',
	sup: 'text-xs align-super',
	sub: 'text-xs align-sub',
	spoiler: 'bg-zinc-500 text-zinc-500 hover:text-white',
};

/**
 * Parse a single line of post text into a markup AST.
 * Unknown / unbalanced tags are left as literal text (no HTML injection path).
 */
export function parseMarkup(input: string): MarkupNode[] {
	let i = 0;

	function parseUntil(endTag: string | null): MarkupNode[] {
		const nodes: MarkupNode[] = [];
		let text = '';

		const flush = () => {
			if (text) {
				nodes.push({ type: 'text', value: text });
				text = '';
			}
		};

		while (i < input.length) {
			if (input[i] !== '[') {
				text += input[i];
				i += 1;
				continue;
			}

			const close = input.indexOf(']', i);
			if (close === -1) {
				text += input.slice(i);
				i = input.length;
				break;
			}

			const raw = input.slice(i + 1, close);

			if (raw.startsWith('/')) {
				const name = raw.slice(1);
				if (endTag !== null && name === endTag) {
					flush();
					i = close + 1;
					return nodes;
				}
				text += input.slice(i, close + 1);
				i = close + 1;
				continue;
			}

			if (TAG_SET.has(raw)) {
				flush();
				i = close + 1;
				const children = parseUntil(raw);
				nodes.push({ type: 'tag', tag: raw as MarkupTag, children });
				continue;
			}

			text += input[i];
			i += 1;
		}

		flush();
		return nodes;
	}

	return parseUntil(null);
}

export type LineKind = 'reply' | 'greentext' | 'plain';

export function classifyLine(trimmed: string): LineKind {
	if (trimmed.startsWith('>>')) return 'reply';
	if (trimmed.startsWith('>')) return 'greentext';
	return 'plain';
}

/** Same href behavior as the old PostBody (`>>123` → `#123`). */
export function replyHref(trimmed: string): string {
	return `#${trimmed.replaceAll('>>', '')}`;
}

export type PostLine = {
	trimmed: string;
	kind: LineKind;
	nodes: MarkupNode[];
	/** Set when `kind === 'reply'`. */
	href?: string;
};

/** Split a full post into display lines with parsed markup (unit-testable). */
export function buildPostLines(text: string): PostLine[] {
	return text.split('\n').map((line) => {
		const trimmed = line.trim();
		const kind = classifyLine(trimmed);
		const nodes = parseMarkup(trimmed);
		if (kind === 'reply') {
			return { trimmed, kind, nodes, href: replyHref(trimmed) };
		}
		return { trimmed, kind, nodes };
	});
}
