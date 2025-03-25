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
	attachments: null;
	is_author: boolean;
};