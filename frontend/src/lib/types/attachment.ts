export type Attachment = {
	id: string | undefined;
	link: string | undefined;
	post_id: string | undefined;
	name: string;
	type: string;
	body: string;
};

export type File = {
	id: string;
	name: string;
	blob: string | ArrayBuffer | null;
};