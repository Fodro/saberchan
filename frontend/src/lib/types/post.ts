import type { Attachment } from "./attachment";

export type Post = {
	id: string;
	number: number;
	text: string;
	thread_id: string;
	sage: boolean;
	op_marker: boolean;
	browser_fingerprint: string;
	ip: string;
	created_at: string;
	attachments: Attachment[];
	is_author: boolean;
};